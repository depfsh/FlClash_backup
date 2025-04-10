cmake_minimum_required(VERSION 3.22.1)


project("core")
add_library(${CMAKE_PROJECT_NAME} SHARED
        core.cpp)

set(LIB_CLASH_PATH "${CMAKE_SOURCE_DIR}/../jniLibs/${ANDROID_ABI}/libclash.so")

if (EXISTS ${LIB_CLASH_PATH})
    message(STATUS "Found libclash.so for ABI ${ANDROID_ABI}")
    add_definitions(-D_LIBCLASH)
    include_directories(${CMAKE_SOURCE_DIR}/../jniLibs/${ANDROID_ABI})
    add_library(clash SHARED IMPORTED)
    set_target_properties(clash PROPERTIES
            IMPORTED_LOCATION ${LIB_CLASH_PATH}
    )
    target_link_libraries(${CMAKE_PROJECT_NAME}
            android
            log
            clash)
else ()
    target_link_libraries(${CMAKE_PROJECT_NAME}
            android
            log
    )
endif ()