package player

import (
	"bytes"
	"errors"
	"image/gif"
	"image/png"
	"io"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/otiai10/gosseract/v2"
)

var registrationMux = &sync.Mutex{}
var gosseractClientCA *gosseract.Client

func (p *Player) registerPlayer() error {
	// Slow down
	time.Sleep(time.Second)

	captchaCode, err := p.getUnregisteredCaptchaCode()
	if err != nil {
		return err
	}
	p.Log.Debug("Got captcha code:", captchaCode)

	params := url.Values{}
	params.Add("nick", p.Config.Nick)
	params.Add("passs", p.Config.PassPlain)
	params.Add("passs2", p.Config.PassPlain)
	params.Add("ko", captchaCode)
	params.Add("null", "Registruotis")
	body := strings.NewReader(params.Encode())

	// Submit registration form
	resp, err := p.httpRequest("POST", *p.Config.Settings.RootAddress+"/"+"index.php?id=reg2&mo=Human&world=1", body)
	if err != nil {
		p.Log.Warning("Failed to perform registration request:", err)
		return err
	}
	defer resp.Body.Close()

	// Create GoQuery document out of response body
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		p.Log.Warning("Failed to create GoQuery doc out of response body:", err)
		return err
	}

	// Retry if invalid CA code, which is fine
	if doc.Find("div:contains('Blogas kodas')").Length() > 0 {
		p.Log.Debug("Invalid captcha CA code, retrying...")
		return p.registerPlayer()
	}

	// If failed to register
	if doc.Find("div:contains('Jūs užregistruotas sėkmingai')").Length() == 0 {
		return errors.New("failed to register player")
	}

	p.Log.Info("Successfully registered player")

	time.Sleep(time.Second)

	// Submit registration form
	resp, err = p.httpRequest("GET", p.renderFullLink("/zaisti.php?{{ creds }}&tipas=0"), nil)
	if err != nil {
		p.Log.Warning("Failed to perform character warrior type request:", err)
		return err
	}
	defer resp.Body.Close()

	// Create GoQuery document out of response body
	doc, err = goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		p.Log.Warning("Failed to create GoQuery doc out of response body for warrior type:", err)
		return err
	}

	// If failed to choose warior type
	if doc.Find("div:contains('Tipas sėkmingai pasirinktas')").Length() == 0 {
		return errors.New("failed to choose warrior type")
	}

	p.Log.Info("Successfully selected warrior type")

	return nil
}

var reCaDigits = regexp.MustCompile(`[^0-9]`)

// This function attempts to load ca.php digits. MIGHT PROVIDE INCORRECT RESULT, RE-RUN IN SUCH CASE
func (p *Player) getUnregisteredCaptchaCode() (string, error) {
	resp, err := p.httpRequest("GET", *p.Config.Settings.RootAddress+"/"+"ca.php", nil)
	if err != nil {
		p.Log.Warning("Failed to perform registration request:", err)
		return "", err
	}
	defer resp.Body.Close()

	// Download image body
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", errors.New("Failed to read captcha image body: " + err.Error())
	}

	// // Convert static GIF to PNG
	// caPng, err := ConvertGifToPng(content)
	// if err != nil {
	// 	return "", errors.New("Failed to convert ca.php captcha image to PNG: " + err.Error())
	// }

	// // Debug: Write to file
	// os.WriteFile("/tmp/ca.png", caPng, 0644)

	// Read text from image
	err = gosseractClientCA.SetImageFromBytes(content)
	if err != nil {
		return "", errors.New("Failed to set image from bytes: " + err.Error())
	}
	text, err := gosseractClientCA.Text()
	if err != nil {
		return "", errors.New("Failed to read text from ca.php captcha image: " + err.Error())
	}
	caText := strings.ToLower(reCaDigits.ReplaceAllString(text, "")) // Already trimmed by regex

	if len(caText) < 2 {
		return p.getUnregisteredCaptchaCode()
	}

	return caText, nil
}

func ConvertGifToPng(gifBytes []byte) ([]byte, error) {
	// Create a bytes reader from the gif bytes
	gifReader := bytes.NewReader(gifBytes)

	// Decode the gif image
	gifImg, err := gif.Decode(gifReader)
	if err != nil {
		return nil, err
	}

	// Create a buffer to write the png to
	pngBuffer := new(bytes.Buffer)

	// Encode the image to png
	err = png.Encode(pngBuffer, gifImg)
	if err != nil {
		return nil, err
	}

	// Return the bytes of the png image
	return pngBuffer.Bytes(), nil
}

func init() {
	// Init tesseract OCR for ca.php captchas
	gosseractClientCA = gosseract.NewClient()
	gosseractClientCA.SetWhitelist("0123456789")
}
