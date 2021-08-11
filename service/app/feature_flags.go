package app

import (
	"strings"

	"github.com/getbread/breadkit/featureflags"
	"github.com/sirupsen/logrus"
)

func featureWhitelisted(flag, shop string) bool {
	whitelisted := false
	shopList := parseFlagToList(flag)
	for _, v := range shopList {
		if shop == v {
			whitelisted = true
			break
		}
	}
	return whitelisted
}

func parseFlagToList(flag string) []string {
	value := featureflags.GetString(flag, "")
	shopList := strings.Split(value, ";")

	if len(shopList) == 0 {
		logrus.WithFields(logrus.Fields{
			"featureFlag": flag,
			"value":       value,
		}).Error("(MiltonFeatureFlags) Feature flag is empty")
	}

	return shopList
}
