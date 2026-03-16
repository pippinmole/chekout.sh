#import <Foundation/Foundation.h>
#import <AppKit/AppKit.h>

// ---------------------------------------------------------------------------
// Thread-safe URL queue
// ---------------------------------------------------------------------------
static NSMutableArray<NSString *> *gURLQueue = nil;
static NSLock *gURLLock = nil;

// ---------------------------------------------------------------------------
// Objective-C delegate that receives Apple Events
// ---------------------------------------------------------------------------
@interface ChekoutURLHandler : NSObject
- (void)handleGetURLEvent:(NSAppleEventDescriptor *)event
           withReplyEvent:(NSAppleEventDescriptor *)replyEvent;
@end

@implementation ChekoutURLHandler
- (void)handleGetURLEvent:(NSAppleEventDescriptor *)event
           withReplyEvent:(NSAppleEventDescriptor *)replyEvent {
    NSString *urlString = [[event paramDescriptorForKeyword:keyDirectObject]
                           stringValue];
    if (!urlString) return;
    [gURLLock lock];
    [gURLQueue addObject:urlString];
    [gURLLock unlock];
}
@end

static ChekoutURLHandler *gHandler = nil;

// ---------------------------------------------------------------------------
// C functions called from Go via CGO
// ---------------------------------------------------------------------------

void RegisterAppleEventHandler(void) {
    if (!gURLQueue) {
        gURLQueue = [NSMutableArray new];
        gURLLock  = [NSLock new];
    }
    if (!gHandler) {
        gHandler = [ChekoutURLHandler new];
    }

    [[NSAppleEventManager sharedAppleEventManager]
        setEventHandler:gHandler
            andSelector:@selector(handleGetURLEvent:withReplyEvent:)
          forEventClass:kInternetEventClass
             andEventID:kAEGetURL];
}

// PopURL returns the oldest queued URL as a malloc'd C string (caller must free),
// or NULL if the queue is empty.
char *PopURL(void) {
    if (!gURLLock) return NULL;

    [gURLLock lock];
    NSString *url = nil;
    if (gURLQueue.count > 0) {
        url = gURLQueue[0];
        [gURLQueue removeObjectAtIndex:0];
    }
    [gURLLock unlock];

    if (!url) return NULL;
    return strdup([url UTF8String]);
}
