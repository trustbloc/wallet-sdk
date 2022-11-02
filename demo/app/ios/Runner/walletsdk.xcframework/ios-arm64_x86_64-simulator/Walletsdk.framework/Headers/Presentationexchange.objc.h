// Objective-C API for talking to github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/presentationexchange Go package.
//   gobind -lang=objc github.com/trustbloc/wallet-sdk/cmd/wallet-sdk-gomobile/presentationexchange
//
// File is generated by gobind. Do not edit.

#ifndef __Presentationexchange_H__
#define __Presentationexchange_H__

@import Foundation;
#include "ref.h"
#include "Universe.objc.h"


@class PresentationexchangeExchange;

/**
 * An Exchange is used to perform matching operations.
 */
@interface PresentationexchangeExchange : NSObject <goSeqRefInterface> {
}
@property(strong, readonly) _Nonnull id _ref;

- (nonnull instancetype)initWithRef:(_Nonnull id)ref;
/**
 * NewExchange returns a new Exchange object.
 */
- (nullable instancetype)init;
/**
 * Match does something (TODO: Implement).
 */
- (NSData* _Nullable)match:(NSData* _Nullable)presentationDefinition vp:(NSData* _Nullable)vp matchOptions:(NSString* _Nullable)matchOptions error:(NSError* _Nullable* _Nullable)error;
@end

/**
 * NewExchange returns a new Exchange object.
 */
FOUNDATION_EXPORT PresentationexchangeExchange* _Nullable PresentationexchangeNewExchange(void);

#endif
