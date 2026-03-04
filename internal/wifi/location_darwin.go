//go:build darwin && cgo

package wifi

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreLocation -framework Foundation

#import <CoreLocation/CoreLocation.h>
#import <Foundation/Foundation.h>

@interface WifiLocationDelegate : NSObject <CLLocationManagerDelegate>
@property (nonatomic, assign) BOOL responded;
@property (nonatomic, assign) int authStatus;
@end

@implementation WifiLocationDelegate
- (void)locationManagerDidChangeAuthorization:(CLLocationManager *)manager {
	CLAuthorizationStatus status = manager.authorizationStatus;
	self.authStatus = (int)status;
	if (status != kCLAuthorizationStatusNotDetermined) {
		self.responded = YES;
	}
}
@end

// requestLocationAuth triggers the Location Services authorization dialog.
// The permission is attributed to the terminal app hosting the CLI tool.
int requestLocationAuth(int *outStatus) {
	@autoreleasepool {
		CLLocationManager *manager = [[CLLocationManager alloc] init];
		WifiLocationDelegate *delegate = [[WifiLocationDelegate alloc] init];
		manager.delegate = delegate;

		CLAuthorizationStatus current = manager.authorizationStatus;
		if (current != kCLAuthorizationStatusNotDetermined) {
			*outStatus = (int)current;
			return 0;
		}

		// requestWhenInUseAuthorization is the correct API on macOS 11+.
		// startUpdatingLocation does not trigger the permission dialog on modern macOS.
		if (@available(macOS 11.0, *)) {
			[manager requestWhenInUseAuthorization];
		} else {
			[manager requestAlwaysAuthorization];
		}

		// Run the run loop to allow the authorization dialog to appear.
		// Short timeout — CLI tools almost never get the dialog on modern macOS
		// because they lack a proper app bundle with Info.plist keys.
		NSDate *timeout = [NSDate dateWithTimeIntervalSinceNow:3.0];
		while (!delegate.responded &&
			   [[NSDate date] compare:timeout] == NSOrderedAscending) {
			[[NSRunLoop currentRunLoop]
				runUntilDate:[NSDate dateWithTimeIntervalSinceNow:0.5]];
		}

		*outStatus = delegate.authStatus;
		return delegate.responded ? 0 : -1;
	}
}

int getLocationAuthStatus(void) {
	@autoreleasepool {
		CLLocationManager *manager = [[CLLocationManager alloc] init];
		return (int)manager.authorizationStatus;
	}
}
*/
import "C"

import "fmt"

// LocationAuthStatus returns the current Location Services authorization status.
func LocationAuthStatus() int {
	return int(C.getLocationAuthStatus())
}

// RequestLocationPermission triggers the Location Services authorization dialog.
// The permission is attributed to the terminal app (e.g. Ghostty, Terminal.app).
// Returns whether authorization was granted.
func RequestLocationPermission() (bool, error) {
	var status C.int
	ret := C.requestLocationAuth(&status)
	if ret != 0 {
		return false, fmt.Errorf("location permission request timed out")
	}
	s := int(status)
	return s == LocationStatusAuthorized || s == LocationStatusAuthorizedWhenInUse, nil
}
