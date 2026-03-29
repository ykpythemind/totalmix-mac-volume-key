#ifndef KEYTAP_DARWIN_H
#define KEYTAP_DARWIN_H

// Media key types from IOKit/hidsystem/ev_keymap.h
#define NX_KEYTYPE_SOUND_UP   0
#define NX_KEYTYPE_SOUND_DOWN 1
#define NX_KEYTYPE_MUTE       7

// Start the event tap (blocks on CFRunLoop)
void startEventTap(void);

#endif
