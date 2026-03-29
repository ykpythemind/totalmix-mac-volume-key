#include <CoreGraphics/CoreGraphics.h>
#include "keytap_darwin.h"

// Defined in keytap.go via //export
extern void goMediaKeyCallback(int keyCode, int keyDown);

static CFMachPortRef eventTapPort = NULL;

static CGEventRef eventCallback(CGEventTapProxy proxy, CGEventType type, CGEventRef event, void *userInfo) {
    // Re-enable the tap if it gets disabled by timeout
    if (type == kCGEventTapDisabledByTimeout || type == kCGEventTapDisabledByUserInput) {
        if (eventTapPort) {
            CGEventTapEnable(eventTapPort, true);
        }
        return event;
    }

    // We only care about NX_SYSDEFINED events
    if (type != NX_SYSDEFINED) {
        return event;
    }

    // Extract the event data
    // For NX_SYSDEFINED media key events:
    // - Field 85 (NSEventSubtype): must be 8 for media keys
    // - Field 86: packed data containing keyCode and keyState
    int64_t subtype = CGEventGetIntegerValueField(event, 85);
    if (subtype != 8) {
        return event;
    }

    int64_t data = CGEventGetIntegerValueField(event, 86);
    int keyCode = (int)((data & 0xFFFF0000) >> 16);
    int keyDown = (int)(((data & 0xFF00) >> 8) == 0x0A);  // 0x0A = key down, 0x0B = key up

    // Filter for volume/mute keys only
    if (keyCode != NX_KEYTYPE_SOUND_UP &&
        keyCode != NX_KEYTYPE_SOUND_DOWN &&
        keyCode != NX_KEYTYPE_MUTE) {
        return event;
    }

    // Only act on key down events
    if (keyDown) {
        goMediaKeyCallback(keyCode, keyDown);
    }

    // Return NULL to consume the event (prevent system error sound)
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
        // Failed to create event tap - likely missing accessibility permissions
        return;
    }

    CFRunLoopSourceRef runLoopSource = CFMachPortCreateRunLoopSource(kCFAllocatorDefault, eventTapPort, 0);
    CFRunLoopAddSource(CFRunLoopGetCurrent(), runLoopSource, kCFRunLoopCommonModes);
    CGEventTapEnable(eventTapPort, true);

    CFRelease(runLoopSource);

    // This blocks forever
    CFRunLoopRun();
}
