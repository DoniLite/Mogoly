// Copyright 2025 DoniLite. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package core

type DNSServer struct {
	isLocal   func(string) bool
	forwardTo string // optional upstream (ip:port)
}
