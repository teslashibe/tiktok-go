package tiktok

import (
	"hash/crc32"
	"time"
)

// xbogusChars is the custom encoding alphabet used for X-Bogus values.
const xbogusChars = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"

// calcXBogus generates the X-Bogus query parameter for a TikTok API URL.
//
// Algorithm (reverse-engineered from TikTok's JS bundle):
//  1. Build a 19-byte array with magic header, CRC32 of query string,
//     CRC32 of User-Agent, timestamp, and XOR checksum.
//  2. Base64-like encode using xbogusChars.
//  3. Return first 25 characters.
//
// queryStr is the URL query string (the part after '?'), without X-Bogus itself.
func calcXBogus(queryStr, ua string) string {
	queryCRC := crc32.ChecksumIEEE([]byte(queryStr))
	uaCRC := crc32.ChecksumIEEE([]byte(ua))

	ts := uint32(time.Now().Unix())

	arr := [19]byte{
		0x02, 0x01, 0x01, 0x01, // magic header
		0xB8, 0x01,             // version tag
		0x40,                   // flags
		byte(queryCRC >> 24),
		byte(queryCRC >> 16),
		byte(queryCRC >> 8),
		byte(queryCRC),
		byte(uaCRC >> 24),
		byte(uaCRC >> 16),
		byte(uaCRC >> 8),
		byte(uaCRC),
		byte(ts >> 16),
		byte(ts >> 8),
		byte(ts),
		0x00, // checksum placeholder (index 18)
	}

	// XOR checksum of first 18 bytes
	var chk byte
	for i := 0; i < 18; i++ {
		chk ^= arr[i]
	}
	arr[18] = chk

	return encodeXBogus(arr[:])
}

// encodeXBogus applies a base64-like encoding using xbogusChars and returns
// the first 25 characters of the result.
func encodeXBogus(data []byte) string {
	chars := []byte(xbogusChars)
	result := make([]byte, 0, 28)

	for i := 0; i < len(data); i += 3 {
		b0 := int(data[i])
		b1, b2 := 0, 0
		if i+1 < len(data) {
			b1 = int(data[i+1])
		}
		if i+2 < len(data) {
			b2 = int(data[i+2])
		}

		result = append(result,
			chars[(b0>>2)&0x3F],
			chars[((b0&0x03)<<4)|(b1>>4)],
			chars[((b1&0x0F)<<2)|(b2>>6)],
			chars[b2&0x3F],
		)
	}

	if len(result) >= 25 {
		return string(result[:25])
	}
	return string(result)
}

