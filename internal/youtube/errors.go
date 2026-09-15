package youtube

import (
	"errors"
	"strings"

	"google.golang.org/api/googleapi"
)

var (
	// ErrEmptyAPIKey is returned by [New] when the key is missing or whitespace.
	ErrEmptyAPIKey = errors.New("youtube: API key is empty")
	// ErrInvalidAPIKey is returned when YouTube rejects the configured API key.
	// Callers must not start the no-key client in this case.
	ErrInvalidAPIKey = errors.New("youtube: invalid API key")
	// ErrNotLive is returned when the video has no active live chat.
	ErrNotLive = errors.New("youtube: video is not a current live stream with active chat")
	// ErrQuotaExceeded is returned when the Data API quota is exhausted.
	ErrQuotaExceeded = errors.New("youtube: API quota exceeded")
)

// IsInvalidAPIKey reports whether err is or wraps [ErrInvalidAPIKey].
func IsInvalidAPIKey(err error) bool {
	return errors.Is(err, ErrInvalidAPIKey)
}

func mapAPIError(err error) error {
	if err == nil {
		return nil
	}
	var ge *googleapi.Error
	if errors.As(err, &ge) {
		if isQuotaGoogleError(ge) {
			return wrapAPI(ErrQuotaExceeded, err)
		}
		if isInvalidKeyGoogleError(ge) {
			return wrapAPI(ErrInvalidAPIKey, err)
		}
	}
	low := strings.ToLower(err.Error())
	if strings.Contains(low, "invalidapikey") || strings.Contains(low, "api key not valid") || strings.Contains(low, "keyinvalid") {
		return wrapAPI(ErrInvalidAPIKey, err)
	}
	if strings.Contains(low, "quotaexceeded") || strings.Contains(low, "quota exceeded") {
		return wrapAPI(ErrQuotaExceeded, err)
	}
	return err
}

func wrapAPI(sentinel, err error) error {
	return errors.Join(sentinel, err)
}

func isQuotaGoogleError(ge *googleapi.Error) bool {
	if ge.Code == 403 {
		low := strings.ToLower(ge.Message + ge.Body)
		if strings.Contains(low, "quota") {
			return true
		}
		for _, e := range ge.Errors {
			if strings.EqualFold(e.Reason, "quotaExceeded") || strings.EqualFold(e.Reason, "dailyLimitExceeded") {
				return true
			}
		}
	}
	return false
}

func isInvalidKeyGoogleError(ge *googleapi.Error) bool {
	for _, e := range ge.Errors {
		switch strings.ToLower(e.Reason) {
		case "keyinvalid", "autherror", "forbidden":
			if ge.Code == 401 || ge.Code == 400 {
				return true
			}
		}
	}
	low := strings.ToLower(ge.Message + ge.Body)
	if ge.Code == 400 || ge.Code == 401 {
		if strings.Contains(low, "api key") || strings.Contains(low, "apikey") || strings.Contains(low, "keyinvalid") {
			return true
		}
	}
	if ge.Code == 403 && (strings.Contains(low, "access not configured") || strings.Contains(low, "has not been used") || strings.Contains(low, "disabled")) {
		return true
	}
	return ge.Code == 401
}
