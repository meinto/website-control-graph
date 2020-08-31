package main

import (
	"context"
	"io/ioutil"
	"log"

	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
	"github.com/meinto/website-control-graph/chrome"
)

func main() {
	var omitEmty *bool
	c := chrome.New(20, omitEmty)
	ctx, cancel := c.CreateContext()
	defer cancel()

	// see: https://intoli.com/blog/not-possible-to-block-chrome-headless/
	const script = `(function(w, n, wn) {
  // Pass the Webdriver Test.
  Object.defineProperty(n, 'webdriver', {
    get: () => false,
  });

  // Pass the Plugins Length Test.
  // Overwrite the plugins property to use a custom getter.
  Object.defineProperty(n, 'plugins', {
    // This just needs to have length > 0 for the current test,
    // but we could mock the plugins too if necessary.
    get: () => [1, 2, 3, 4, 5],
  });

  // Pass the Languages Test.
  // Overwrite the plugins property to use a custom getter.
  Object.defineProperty(n, 'languages', {
    get: () => ['en-US', 'en'],
  });

  // Pass the Chrome Test.
  // We can mock this in as much depth as we need for the test.
  w.chrome = {
    runtime: {},
  };

  // Pass the Permissions Test.
  const originalQuery = wn.permissions.query;
  return wn.permissions.query = (parameters) => (
    parameters.name === 'notifications' ?
      Promise.resolve({ state: Notification.permission }) :
      originalQuery(parameters)
  );

})(window, navigator, window.navigator);`

	var buf []byte
	var scriptID page.ScriptIdentifier
	if err := chromedp.Run(
		ctx,
		//chromedp.Emulate(device.IPhone7),
		chromedp.ActionFunc(func(ctx context.Context) error {
			var err error
			scriptID, err = page.AddScriptToEvaluateOnNewDocument(script).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
		chromedp.Navigate("https://intoli.com/blog/not-possible-to-block-chrome-headless/chrome-headless-test.html"),
		chromedp.CaptureScreenshot(&buf),
	); err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile("screenshot.png", buf, 0644); err != nil {
		log.Fatal(err)
	}
}
