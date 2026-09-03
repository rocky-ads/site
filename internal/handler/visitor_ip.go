package handler

import (
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// VisitorIP returns the public client address when the request arrived
// through Render/Cloudflare. CF-Connecting-IP and True-Client-IP are
// set by the edge and are preferred over X-Forwarded-For, which Render
// appends to and does not strip. Local and internal probes fall back
// to the TCP peer.
func VisitorIP(c *fiber.Ctx) string {
	if ip := publicIP(c.Get("CF-Connecting-IP")); ip != "" {
		return ip
	}
	if ip := publicIP(c.Get("True-Client-IP")); ip != "" {
		return ip
	}
	if ip := firstPublicForwarded(
		c.Get(fiber.HeaderXForwardedFor)); ip != "" {
		return ip
	}
	return c.IP()
}

func firstPublicForwarded(header string) string {
	for _, part := range strings.Split(header, ",") {
		if ip := publicIP(part); ip != "" {
			return ip
		}
	}
	return ""
}

func publicIP(s string) string {
	ip := net.ParseIP(strings.TrimSpace(s))
	if ip == nil || !isPublicIP(ip) {
		return ""
	}
	return ip.String()
}

func isPublicIP(ip net.IP) bool {
	return !ip.IsPrivate() && !ip.IsLoopback() &&
		!ip.IsLinkLocalUnicast() && !ip.IsLinkLocalMulticast() &&
		!ip.IsMulticast() && !ip.IsUnspecified()
}
