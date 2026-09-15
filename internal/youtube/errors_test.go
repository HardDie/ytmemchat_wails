package youtube

import (
	"errors"
	"testing"

	"google.golang.org/api/googleapi"
)

func TestIsInvalidAPIKey_googleapi(t *testing.T) {
	err := mapAPIError(&googleapi.Error{
		Code:    400,
		Message: "API key not valid. Please pass a valid API key.",
		Errors:  []googleapi.ErrorItem{{Reason: "keyInvalid"}},
	})
	if !IsInvalidAPIKey(err) {
		t.Fatalf("%v", err)
	}
	quota := mapAPIError(&googleapi.Error{
		Code:    403,
		Message: "The request cannot be completed because you have exceeded your quota.",
		Errors:  []googleapi.ErrorItem{{Reason: "quotaExceeded"}},
	})
	if IsInvalidAPIKey(quota) {
		t.Fatal("quota classified as invalid key")
	}
	if !errors.Is(quota, ErrQuotaExceeded) {
		t.Fatalf("quota = %v", quota)
	}
	notLive := mapAPIError(ErrNotLive)
	if IsInvalidAPIKey(notLive) {
		t.Fatal("not live")
	}
}
