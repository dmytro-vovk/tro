package app

import qr "github.com/skip2/go-qrcode"

type QRRequest struct {
	Data string `json:"data"`
}

func (Application) QR(r *QRRequest) ([]byte, error) {
	code, err := qr.New(r.Data, qr.Low)
	if err != nil {
		return nil, err
	}

	return code.PNG(400)
}
