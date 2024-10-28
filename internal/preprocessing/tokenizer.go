package preprocessing

import (
	"regexp"
	"strings"
	"unicode"
)

// Tokenizer defines the methods for tokenizing and preprocessing text
type Tokenizer struct {
	removePunctuation bool
	lowercase         bool
	removeStopWords   bool
	stopWords         map[string]struct{}
}

// NewTokenizer initializes a Tokenizer with customizable options
func NewTokenizer(removePunctuation, lowercase, removeStopWords bool, stopWords []string) *Tokenizer {
	stopWordsMap := make(map[string]struct{})
	for _, word := range stopWords {
		stopWordsMap[word] = struct{}{}
	}
	return &Tokenizer{
		removePunctuation: removePunctuation,
		lowercase:         lowercase,
		removeStopWords:   removeStopWords,
		stopWords:         stopWordsMap,
	}
}

// Tokenize splits the text into tokens based on whitespace
func (t *Tokenizer) Tokenize(text string) []string {
	if t.lowercase {
		text = strings.ToLower(text)
	}
	if t.removePunctuation {
		text = removePunctuation(text)
	}

	tokens := strings.Fields(text)

	if t.removeStopWords {
		tokens = t.filterStopWords(tokens)
	}

	return tokens
}

// filterStopWords removes common stop words from tokens
func (t *Tokenizer) filterStopWords(tokens []string) []string {
	var filtered []string
	for _, token := range tokens {
		if _, found := t.stopWords[token]; !found {
			filtered = append(filtered, token)
		}
	}
	return filtered
}

// removePunctuation removes punctuation from the input text
func removePunctuation(text string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsPunct(r) {
			return -1
		}
		return r
	}, text)
}

// NGram generates n-grams from tokens
func (t *Tokenizer) NGram(tokens []string, n int) [][]string {
	var ngrams [][]string
	for i := 0; i <= len(tokens)-n; i++ {
		ngrams = append(ngrams, tokens[i:i+n])
	}
	return ngrams
}

// Advanced Preprocessing (optional): Removes URLs, digits, or other specific patterns
func (t *Tokenizer) CleanText(text string) string {
	// Remove URLs
	reURL := regexp.MustCompile(`http[s]?://\S+`)
	text = reURL.ReplaceAllString(text, "")

	// Remove digits
	reDigits := regexp.MustCompile(`\d+`)
	text = reDigits.ReplaceAllString(text, "")

	return strings.TrimSpace(text)
}
