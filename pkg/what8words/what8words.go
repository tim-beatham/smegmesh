// Package to convert an IPV6 addres into 8 words
package what8words

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strings"
)

type What8Words struct {
	words []string
}

// Convert implements What8Words.
func (w *What8Words) Convert(ipStr string) (string, error) {
	ip, ipNet, err := net.ParseCIDR(ipStr)

	if err != nil {
		return "", err
	}

	ip16 := ip.To16()

	if ip16 == nil {
		return "", fmt.Errorf("cannot convert ip to 16 representation")
	}

	representation := make([]string, 7)

	for i := 2; i <= net.IPv6len-2; i += 2 {
		word1 := w.words[ip16[i]]
		word2 := w.words[ip16[i+1]]

		representation[i/2-1] = fmt.Sprintf("%s-%s", word1, word2)
	}

	prefixSize, _ := ipNet.Mask.Size()
	return strings.Join(representation[:prefixSize/16-1], "."), nil
}

// Convert implements What8Words.
func (w *What8Words) ConvertIdentifier(ipStr string) (string, error) {
	ip, err := w.Convert(ipStr)

	if err != nil {
		return "", err
	}

	constituents := strings.Split(ip, ".")

	return strings.Join(constituents[3:], "."), nil
}
func NewWhat8Words(pathToWords string) (*What8Words, error) {
	words, err := ReadWords(pathToWords)

	if err != nil {
		return nil, err
	}

	return &What8Words{words: words}, nil
}

// ReadWords reads the what 8 words txt file
func ReadWords(wordFile string) ([]string, error) {
	f, err := os.ReadFile(wordFile)

	if err != nil {
		return nil, err
	}

	words := make([]string, 257)

	reader := bufio.NewScanner(bytes.NewReader(f))

	counter := 0

	for reader.Scan() && counter <= len(words) {
		text := reader.Text()

		words[counter] = text
		counter++

		if reader.Err() != nil {
			return nil, reader.Err()
		}
	}

	return words, nil
}
