package vo

import (
	"fmt"
	"strings"

	"github.com/D1sordxr/image-processor/internal/domain/core/shared/vo"
)

type ResultUrl string

const resultUrlSuffix = "/image/"

func (r ResultUrl) String() string {
	return string(r)
}

func (r ResultUrl) IsValid() bool {
	return strings.Contains(r.String(), resultUrlSuffix)
}

func NewResultUrl(baseURL vo.BaseURL, id string) ResultUrl {
	return ResultUrl(fmt.Sprintf("%s/%s/%s", baseURL.String(), "image", id))
}

func NewValidResultUrl(s string) (ResultUrl, error) {
	url := ResultUrl(s)
	if !url.IsValid() {
		return "", fmt.Errorf("invalid result url: %s", s)
	}
	return url, nil
}
