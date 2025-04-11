android_arm64:
	dart ./setup.dart android --arch arm64 --out core
	flutter run

android_arm64_core:
	dart ./setup.dart android --arch arm64 --out core

macos_arm64:
	dart ./setup.dart android --arch arm64