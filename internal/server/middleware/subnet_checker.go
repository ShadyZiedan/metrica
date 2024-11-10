package middleware

import (
	"fmt"
	"net"
	"net/http"
)

func trustedSubNetCheck(trustedSubnet *net.IPNet) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			agentIP := r.Header.Get("X-Real-IP")
			if agentIP == "" {
				http.Error(w, "Missing X-Real-IP header", http.StatusForbidden)
				return
			}

			ip := net.ParseIP(agentIP)
			if ip == nil || !trustedSubnet.Contains(ip) {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func NewTrustedSubNetMiddleware(trustedSubnet string) (func(next http.Handler) http.Handler, error) {
	_, trustedNet, err := net.ParseCIDR(trustedSubnet)
	if err != nil {
		return nil, fmt.Errorf("failed to parse trusted subnet: %w", err)
	}
	return trustedSubNetCheck(trustedNet), nil
}
