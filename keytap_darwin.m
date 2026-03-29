#import <Cocoa/Cocoa.h>
#import <CoreGraphics/CoreGraphics.h>
#import <CoreAudio/CoreAudio.h>
#import <stdio.h>
#import "keytap_darwin.h"

// Defined in keytap.go via //export
extern void goMediaKeyCallback(int keyCode, int keyDown);

static CFMachPortRef eventTapPort = NULL;

// Check if the current default output device is an RME device
static int isRMEOutputDevice(void) {
    AudioObjectPropertyAddress propAddr = {
        kAudioHardwarePropertyDefaultOutputDevice,
        kAudioObjectPropertyScopeGlobal,
        kAudioObjectPropertyElementMain
    };

    AudioDeviceID deviceID = kAudioObjectUnknown;
    UInt32 size = sizeof(deviceID);
    OSStatus status = AudioObjectGetPropertyData(
        kAudioObjectSystemObject, &propAddr, 0, NULL, &size, &deviceID
    );
    if (status != noErr || deviceID == kAudioObjectUnknown) {
        return 0;
    }

    // Get the device manufacturer
    propAddr.mSelector = kAudioDevicePropertyDeviceManufacturerCFString;
    propAddr.mScope = kAudioObjectPropertyScopeOutput;
    CFStringRef manufacturer = NULL;
    size = sizeof(manufacturer);
    status = AudioObjectGetPropertyData(deviceID, &propAddr, 0, NULL, &size, &manufacturer);
    if (status == noErr && manufacturer != NULL) {
        // Check for RME in manufacturer string
        CFRange range = CFStringFind(manufacturer, CFSTR("RME"), kCFCompareCaseInsensitive);
        if (range.location != kCFNotFound) {
            CFRelease(manufacturer);
            return 1;
        }

        // Also try "Fireface"
        range = CFStringFind(manufacturer, CFSTR("Fireface"), kCFCompareCaseInsensitive);
        if (range.location != kCFNotFound) {
            CFRelease(manufacturer);
            return 1;
        }
        CFRelease(manufacturer);
    }

    // Also check device name
    propAddr.mSelector = kAudioDevicePropertyDeviceNameCFString;
    CFStringRef deviceName = NULL;
    size = sizeof(deviceName);
    status = AudioObjectGetPropertyData(deviceID, &propAddr, 0, NULL, &size, &deviceName);
    if (status == noErr && deviceName != NULL) {
        CFRange range = CFStringFind(deviceName, CFSTR("RME"), kCFCompareCaseInsensitive);
        if (range.location != kCFNotFound) {
            CFRelease(deviceName);
            return 1;
        }
        range = CFStringFind(deviceName, CFSTR("Fireface"), kCFCompareCaseInsensitive);
        if (range.location != kCFNotFound) {
            CFRelease(deviceName);
            return 1;
        }
        CFRelease(deviceName);
    }

    return 0;
}

static CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *userInfo) {
    // Re-enable the tap if it gets disabled by timeout
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        if (eventTapPort) {
            CGEventTapEnable(eventTapPort, true);
        }
        return event;
    }

    if (type != NX_SYSDEFINED) {
        return event;
    }

    // Convert CGEvent to NSEvent to properly access data1
    NSEvent *nsEvent = [NSEvent eventWithCGEvent:event];
    if (nsEvent == nil) {
        return event;
    }

    // Media key events have subtype 8 (NSEventSubtypeScreenChanged on older APIs,
    // but actually it's the raw subtype value 8)
    if ([nsEvent subtype] != 8) {
        return event;
    }

    NSInteger data1 = [nsEvent data1];

    // data1 layout for media keys:
    //   bits 16-31: key code
    //   bits 8-15:  key flags (0x0A = key down, 0x0B = key up, 0x0A with repeat flag)
    //   bits 0-7:   reserved
    int keyCode = (int)((data1 & 0xFFFF0000) >> 16);
    int keyFlags = (int)((data1 & 0x0000FF00) >> 8);
    int keyDown = (keyFlags & 0x0F) == 0x0A;  // 0x0A = down, 0x0B = up

    // Filter for volume/mute keys only
    if (keyCode != NX_KEYTYPE_SOUND_UP &&
        keyCode != NX_KEYTYPE_SOUND_DOWN &&
        keyCode != NX_KEYTYPE_MUTE) {
        return event;
    }

    // Only intercept when output device is RME; otherwise let system handle it
    if (!isRMEOutputDevice()) {
        return event;
    }

    // Only act on key down events
    if (keyDown) {
        goMediaKeyCallback(keyCode, keyDown);
    }

    // Return NULL to consume the event
    return NULL;
}

void startEventTap(void) {
    CGEventMask eventMask = (1 << NX_SYSDEFINED);

    eventTapPort = CGEventTapCreate(
        kCGSessionEventTap,
        kCGHeadInsertEventTap,
        kCGEventTapOptionDefault,
        eventMask,
        eventCallback,
        NULL
    );

    if (!eventTapPort) {
        fprintf(stderr, "Error: Failed to create event tap. Grant Accessibility permission in System Settings.\n");
        return;
    }
    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTapPort, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTapPort, true);

    CFRelease(runLoopSource);

    CFRunLoopRun();
}
