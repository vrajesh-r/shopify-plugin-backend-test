package app

import (
	"crypto/sha1"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/getbread/shopify_plugin_backend/service/shopify"
	"github.com/getbread/shopify_plugin_backend/service/types"
)

func generateScriptSource(shopID, version string) string {
	cartJs := "cart.js"
	if version == BreadPlatform {
		cartJs = "cart_platform.js"
	}
	return fmt.Sprintf("%s/static/%s/%s", appConfig.HostConfig.MiltonHost, shopID, cartJs)
}

func generateScriptSourceWithHash(shopID, version string, fileHash string) string {
	var queryString string
	if len(fileHash) > 0 {
		queryString = fmt.Sprintf("?v=%s", fileHash)
	}
	format := "%s/static/%s/cart.js%s"
	if version == BreadPlatform {
		format = "%s/static/%s/cart_platform.js%s"
	}
	// If fileHash is empty, caller should expect the base asset URL
	return fmt.Sprintf(format, appConfig.HostConfig.MiltonHost, shopID, queryString)
}

func embedScriptTag(shop types.Shop) error {
	// construct Shopify client
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)

	//Delete script of inactive version
	inactiveVersion := BreadPlatform
	if shop.ActiveVersion == BreadPlatform {
		inactiveVersion = BreadClassic
	}
	go func() {
		removeEmbeddedScripts(shop, inactiveVersion)
	}()

	// construct script tag src
	scriptSrc := generateScriptSource(string(shop.Id), shop.ActiveVersion)

	// check if script already embedded
	scripts, err := queryEmbeddedScriptsBySrc(scriptSrc, sc)
	if err == nil && len(scripts) > 0 { // short circuit if already embedded
		return nil
	}

	// embed scripts
	embedScriptReq := &shopify.EmbedScriptRequest{
		ScriptTag: shopify.ScriptTag{
			Event: "onload",
			Src:   scriptSrc,
		},
	}
	var embedScriptRes shopify.EmbedScriptResponse
	return sc.EmbedScript(embedScriptReq, &embedScriptRes)
}

func embedScriptTagFromVersion(shop types.Shop, req updateVersionRequest) error {
	if shop.ManualEmbedScript {
		return nil
	}

	// construct Shopify client
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)

	//Delete script of inactive version
	inActiveVersion := BreadPlatform
	if req.ActiveVersion == BreadPlatform {
		inActiveVersion = BreadClassic
	}
	go func() {
		removeEmbeddedScripts(shop, inActiveVersion)
	}()

	// construct script tag src
	scriptSrc := generateScriptSource(string(shop.Id), req.ActiveVersion)

	// check if script already embedded
	scripts, err := queryEmbeddedScriptsBySrc(scriptSrc, sc)
	if err == nil && len(scripts) > 0 { // short circuit if already embedded
		return nil
	}

	// embed scripts
	embedScriptReq := &shopify.EmbedScriptRequest{
		ScriptTag: shopify.ScriptTag{
			Event: "onload",
			Src:   scriptSrc,
		},
	}
	var embedScriptRes shopify.EmbedScriptResponse
	return sc.EmbedScript(embedScriptReq, &embedScriptRes)
}

func embedScriptTagWithHash(shop types.Shop, fileHash string) error {
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)
	scriptSrc := generateScriptSourceWithHash(string(shop.Id), shop.ActiveVersion, fileHash)

	// short circuit if script is already embedded
	scripts, err := queryEmbeddedScriptsBySrc(scriptSrc, sc)
	if err == nil && len(scripts) > 0 {
		return nil
	}

	embedScriptReq := &shopify.EmbedScriptRequest{
		ScriptTag: shopify.ScriptTag{
			Event: "onload",
			Src:   scriptSrc,
		},
	}
	var embedScriptRes shopify.EmbedScriptResponse
	return sc.EmbedScript(embedScriptReq, &embedScriptRes)
}

func removeEmbeddedScripts(shop types.Shop, version string) error {
	// construct Shopify client
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)

	// construct script tag src
	scriptSrc := generateScriptSource(string(shop.Id), version)

	// get embedded scripts
	scripts, err := queryEmbeddedScriptsBySrc(scriptSrc, sc)
	if err != nil {
		return err
	}
	if len(scripts) == 0 {
		return nil
	}

	// remove embedded scripts
	var errors []error
	for _, script := range scripts {
		sID := strconv.Itoa(script.Id)
		var res shopify.DeleteEmbeddedScriptResponse
		if err := sc.DeleteEmbeddedScript(sID, &res); err != nil {
			errors = append(errors, err)
		}
	}
	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func removeAllEmbeddedScripts(shop types.Shop) error {
	sc := shopify.NewClient(shop.Shop, shop.AccessToken)

	// construct base script URL for comparison
	baseScriptURL := generateScriptSourceWithHash(string(shop.Id), BreadClassic, "")
	basePlatformScriptURL := generateScriptSourceWithHash(string(shop.Id), BreadPlatform, "")

	// get all embedded scripts
	var res shopify.SearchEmbeddedScriptResponse
	if err := sc.QueryEmbeddedScripts(url.Values{}, &res); err != nil {
		return err
	}

	if len(res.ScriptTags) == 0 {
		return nil
	}

	// remove any Bread embedded scripts matching the base URL
	var errors []error
	for _, script := range res.ScriptTags {
		if strings.Contains(script.Src, baseScriptURL) || strings.Contains(script.Src, basePlatformScriptURL) {
			scriptID := strconv.Itoa(script.Id)
			var res shopify.DeleteEmbeddedScriptResponse
			err := sc.DeleteEmbeddedScript(scriptID, &res)
			if err != nil {
				errors = append(errors, err)
			}
		}
	}

	if len(errors) > 0 {
		return errors[0]
	}
	return nil
}

func queryEmbeddedScriptsBySrc(scriptSrc string, sc *shopify.Client) ([]shopify.ScriptTag, error) {
	v := url.Values{}
	v.Set("src", scriptSrc)

	var res shopify.SearchEmbeddedScriptResponse
	if err := sc.QueryEmbeddedScripts(v, &res); err != nil {
		return nil, err
	}

	return res.ScriptTags, nil
}

func createFileHash(data interface{}) (string, error) {

	input := strings.NewReader(fmt.Sprintf("%v", data))
	hash := sha1.New()

	if _, err := io.Copy(hash, input); err != nil {
		return "", err
	}

	fileHash := fmt.Sprintf("%x", hash.Sum(nil))
	return fileHash[len(fileHash)-8:], nil
}

func updateEmbeddedScriptFromSettings(shop types.Shop, req updateSettingsRequest) error {
	// Remove embedded scripts
	if !shop.ManualEmbedScript {
		if err := removeAllEmbeddedScripts(shop); err != nil {
			return err
		}
	}
	if !req.ManualEmbedScript {
		// Generate file hash for script
		req.BreadSecretKey = ""
		req.BreadSandboxSecretKey = ""
		hash, err := createFileHash(req)
		if err != nil {
			return err
		}

		// Update embedded script
		if err := embedScriptTagWithHash(shop, hash); err != nil {
			return err
		}
	}

	return nil
}

func updateEmbeddedScriptFromVersion(shop types.Shop, req updateVersionRequest) error {
	// Remove embedded scripts
	if !shop.ManualEmbedScript {
		if err := removeAllEmbeddedScripts(shop); err != nil {
			return err
		}

		// Generate file hash for script
		hash, err := createFileHash(req)
		if err != nil {
			return err
		}

		// Update embedded script
		shop.ActiveVersion = req.ActiveVersion
		if err := embedScriptTagWithHash(shop, hash); err != nil {
			return err
		}
	}

	return nil
}
