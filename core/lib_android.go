//go:build android && cgo

package main

/*
extern int myVar;
*/
import "C"
import (
	"context"
	abridge "core/android-bride"
	bridge "core/dart-bridge"
	"core/platform"
	"core/state"
	t "core/tun"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/metacubex/mihomo/common/utils"
	"github.com/metacubex/mihomo/component/dialer"
	"github.com/metacubex/mihomo/component/process"
	"github.com/metacubex/mihomo/constant"
	"github.com/metacubex/mihomo/dns"
	"github.com/metacubex/mihomo/listener/sing_tun"
	"github.com/metacubex/mihomo/log"
	"golang.org/x/sync/semaphore"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"
)

type TunHandler struct {
	listener *sing_tun.Listener
	callback unsafe.Pointer

	limit *semaphore.Weighted
}

func (t *TunHandler) init() {
	initTunHook()
}

func (t *TunHandler) close() {
	_ = t.limit.Acquire(context.TODO(), 4)
	defer t.limit.Release(4)

	removeTunHook()
	if t.listener != nil {
		_ = t.listener.Close()
	}
	t.listener = nil
}

func (t *TunHandler) markSocket(fd int) {
	_ = t.limit.Acquire(context.Background(), 1)
	defer t.limit.Release(1)

	if t.listener == nil {
		return
	}

	abridge.MarkSocket(t.callback, fd)
}

func (t *TunHandler) querySocketUid(protocol int, source, target string) int {
	_ = t.limit.Acquire(context.Background(), 1)
	defer t.limit.Release(1)

	if t.listener == nil {
		return -1
	}

	return abridge.QuerySocketUid(t.callback, protocol, source, target)
}

type Fd struct {
	Id    string `json:"id"`
	Value int64  `json:"value"`
}

type Process struct {
	Id       string             `json:"id"`
	Metadata *constant.Metadata `json:"metadata"`
}

type ProcessMapItem struct {
	Id    string `json:"id"`
	Value string `json:"value"`
}

type InvokeManager struct {
	invokeMap sync.Map
	chanMap   map[string]chan struct{}
	chanLock  sync.Mutex
}

func NewInvokeManager() *InvokeManager {
	return &InvokeManager{
		chanMap: make(map[string]chan struct{}),
	}
}

func (m *InvokeManager) completer(id string, value string) {
	m.invokeMap.Store(id, value)
	m.chanLock.Lock()
	if ch, ok := m.chanMap[id]; ok {
		close(ch)
		delete(m.chanMap, id)
	}
	m.chanLock.Unlock()
}

func (m *InvokeManager) await(id string) string {
	m.chanLock.Lock()
	if _, ok := m.chanMap[id]; !ok {
		m.chanMap[id] = make(chan struct{})
	}
	ch := m.chanMap[id]
	m.chanLock.Unlock()

	timeout := time.After(500 * time.Millisecond)
	select {
	case <-ch:
		res, ok := m.invokeMap.Load(id)
		m.invokeMap.Delete(id)
		if ok {
			return res.(string)
		} else {
			return ""
		}
	case <-timeout:
		m.completer(id, "")
		return ""
	}
}

var (
	invokePort       int64 = -1
	fdInvokeMap            = NewInvokeManager()
	processInvokeMap       = NewInvokeManager()
	tunLock          sync.Mutex
	runTime          *time.Time
	errBlocked       = errors.New("blocked")
	tunHandler       TunHandler
)

func handleStopTun() {
	tunLock.Lock()
	defer tunLock.Unlock()
	runTime = nil
	tunHandler.close()
}

func handleGetRunTime() string {
	if runTime == nil {
		return ""
	}
	return strconv.FormatInt(runTime.UnixMilli(), 10)
}

func handleSetProcessMap(params string) {
	var processMapItem = &ProcessMapItem{}
	err := json.Unmarshal([]byte(params), processMapItem)
	if err == nil {
		processInvokeMap.completer(processMapItem.Id, processMapItem.Value)
	}
}

//export attachInvokePort
func attachInvokePort(mPort C.longlong) {
	invokePort = int64(mPort)
}

func sendInvokeMessage(message InvokeMessage) {
	if invokePort == -1 {
		return
	}
	bridge.SendToPort(invokePort, message.Json())
}

func handleMarkSocket(fd Fd) {
	sendInvokeMessage(InvokeMessage{
		Type: ProtectInvoke,
		Data: fd,
	})
}

func handleParseProcess(process Process) {
	sendInvokeMessage(InvokeMessage{
		Type: ProcessInvoke,
		Data: process,
	})
}

func handleSetFdMap(id string) {
	go func() {
		fdInvokeMap.completer(id, "")
	}()
}

func initTunHook() {
	dialer.DefaultSocketHook = func(network, address string, conn syscall.RawConn) error {
		if platform.ShouldBlockConnection() {
			return errBlocked
		}
		return conn.Control(func(fd uintptr) {
			fdInt := int64(fd)
			id := utils.NewUUIDV1().String()

			handleMarkSocket(Fd{
				Id:    id,
				Value: fdInt,
			})

			fdInvokeMap.await(id)
		})
	}
	process.DefaultPackageNameResolver = func(metadata *constant.Metadata) (string, error) {
		if metadata == nil {
			return "", process.ErrInvalidNetwork
		}
		id := utils.NewUUIDV1().String()
		handleParseProcess(Process{
			Id:       id,
			Metadata: metadata,
		})
		return processInvokeMap.await(id), nil
	}
}

func removeTunHook() {
	dialer.DefaultSocketHook = nil
	process.DefaultPackageNameResolver = nil
}

func handleGetAndroidVpnOptions() string {
	tunLock.Lock()
	defer tunLock.Unlock()
	options := state.AndroidVpnOptions{
		Enable:           state.CurrentState.VpnProps.Enable,
		Port:             currentConfig.General.MixedPort,
		Ipv4Address:      state.DefaultIpv4Address,
		Ipv6Address:      state.GetIpv6Address(),
		AccessControl:    state.CurrentState.VpnProps.AccessControl,
		SystemProxy:      state.CurrentState.VpnProps.SystemProxy,
		AllowBypass:      state.CurrentState.VpnProps.AllowBypass,
		RouteAddress:     currentConfig.General.Tun.RouteAddress,
		BypassDomain:     state.CurrentState.BypassDomain,
		DnsServerAddress: state.GetDnsServerAddress(),
	}
	data, err := json.Marshal(options)
	if err != nil {
		fmt.Println("Error:", err)
		return ""
	}
	return string(data)
}

func handleUpdateDns(value string) {
	go func() {
		log.Infoln("[DNS] updateDns %s", value)
		dns.UpdateSystemDNS(strings.Split(value, ","))
		dns.FlushCacheWithDefaultResolver()
	}()
}

func handleGetCurrentProfileName() string {
	if state.CurrentState == nil {
		return ""
	}
	return state.CurrentState.CurrentProfileName
}

func nextHandle(action *Action, result func(data interface{})) bool {
	switch action.Method {
	case getAndroidVpnOptionsMethod:
		result(handleGetAndroidVpnOptions())
		return true
	case updateDnsMethod:
		data := action.Data.(string)
		handleUpdateDns(data)
		result(true)
		return true
	case setFdMapMethod:
		fdId := action.Data.(string)
		handleSetFdMap(fdId)
		result(true)
		return true
	case setProcessMapMethod:
		data := action.Data.(string)
		handleSetProcessMap(data)
		result(true)
		return true
	case getRunTimeMethod:
		result(handleGetRunTime())
		return true
	case getCurrentProfileNameMethod:
		result(handleGetCurrentProfileName())
		return true
	}
	return false
}

//export quickStart
func quickStart(dirChar *C.char, paramsChar *C.char, stateParamsChar *C.char, port C.longlong) {
	i := int64(port)
	dir := C.GoString(dirChar)
	bytes := []byte(C.GoString(paramsChar))
	stateParams := C.GoString(stateParamsChar)
	go func() {
		res := handleInitClash(dir)
		if res == false {
			bridge.SendToPort(i, "init error")
		}
		handleSetState(stateParams)
		bridge.SendToPort(i, handleUpdateConfig(bytes))
	}()
}

//export startTUN
func startTUN(fd C.int, callback unsafe.Pointer) bool {
	handleStopTun()
	tunLock.Lock()
	defer tunLock.Unlock()
	f := int(fd)
	if f == 0 {
		now := time.Now()
		runTime = &now
	} else {
		tunHandler = TunHandler{
			callback: callback,
		}
		tunHandler.init()
		tunListener, _ := t.Start(f, currentConfig.General.Tun.Device, currentConfig.General.Tun.Stack)
		if tunListener != nil {
			log.Infoln("TUN address: %v", tunListener.Address())
		} else {
			tunHandler.close()
			return false
		}
		tunHandler.listener = tunListener
		now := time.Now()
		runTime = &now
	}
	return true
}

//export getRunTime
func getRunTime() *C.char {
	return C.CString(handleGetRunTime())
}

//export stopTun
func stopTun() {
	handleStopTun()
}

//export setFdMap
func setFdMap(fdIdChar *C.char) {
	fdId := C.GoString(fdIdChar)
	handleSetFdMap(fdId)
}

//export getCurrentProfileName
func getCurrentProfileName() *C.char {
	return C.CString(handleGetCurrentProfileName())
}

//export getAndroidVpnOptions
func getAndroidVpnOptions() *C.char {
	return C.CString(handleGetAndroidVpnOptions())
}

//export setState
func setState(s *C.char) {
	paramsString := C.GoString(s)
	handleSetState(paramsString)
}

//export updateDns
func updateDns(s *C.char) {
	dnsList := C.GoString(s)
	handleUpdateDns(dnsList)
}

//export setProcessMap
func setProcessMap(s *C.char) {
	if s == nil {
		return
	}
	paramsString := C.GoString(s)
	handleSetProcessMap(paramsString)
}
