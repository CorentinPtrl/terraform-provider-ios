// Copyright (c) Corentin Pitrel
// SPDX-License-Identifier: MIT

package utils

import (
	"fmt"
	"net"
	"strconv"
	"strings"
)

func MaskToWildcard(maskStr string) (string, error) {
	mask := net.ParseIP(maskStr)
	if mask == nil {
		return "", fmt.Errorf("invalid IP address: %s", maskStr)
	}
	mask = mask.To4()
	if mask == nil {
		return "", fmt.Errorf("not a valid IPv4 address: %s", maskStr)
	}

	var wildcardParts []string
	for _, b := range mask {
		wildcard := 255 - int(b)
		wildcardParts = append(wildcardParts, fmt.Sprintf("%d", wildcard))
	}

	return strings.Join(wildcardParts, "."), nil
}

func SubnetMaskToCIDR(maskStr string) (int, error) {
	ip := net.ParseIP(maskStr)
	if ip == nil {
		return 0, fmt.Errorf("invalid IP mask string: %s", maskStr)
	}

	ip = ip.To4()
	if ip == nil {
		return 0, fmt.Errorf("not a valid IPv4 mask: %s", maskStr)
	}

	mask := net.IPv4Mask(ip[0], ip[1], ip[2], ip[3])
	ones, _ := mask.Size()
	return ones, nil
}

func WildcardToMask(wildcard string) (net.IPMask, error) {
	octets := strings.Split(wildcard, ".")
	if len(octets) != 4 {
		return nil, fmt.Errorf("invalid wildcard mask: %s", wildcard)
	}

	mask := make(net.IPMask, 4)
	for i, part := range octets {
		val, err := strconv.Atoi(part)
		if err != nil || val < 0 || val > 255 {
			return nil, fmt.Errorf("invalid octet in wildcard: %s", part)
		}
		mask[i] = byte(255 - val)
	}
	return mask, nil
}

func maskToCIDR(mask net.IPMask) (int, error) {
	ones, bits := mask.Size()
	if bits != 32 {
		return 0, fmt.Errorf("not a valid IPv4 mask")
	}
	return ones, nil
}

func WildcardToCIDR(wildcard string) (int, error) {
	mask, err := WildcardToMask(wildcard)
	if err != nil {
		return 0, err
	}
	return maskToCIDR(mask)
}
