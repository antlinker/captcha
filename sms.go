// Copyright 2011-2014 Dmitry Chestnykh. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package captcha

// SMSSender 短信验证码发送
type SMSSender interface {
	Send(tel, code string) error
}
