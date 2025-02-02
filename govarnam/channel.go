package govarnam

import (
	"context"
	"fmt"
	"log"
	"time"
)

type channelDictionaryResult struct {
	exactMatches []Suggestion
	suggestions  []Suggestion
}

func (varnam *Varnam) channelTokenizeWord(ctx context.Context, word string, matchType int, partial bool, channel chan *[]Token) {
	select {
	case <-ctx.Done():
		close(channel)
		return
	default:
		start := time.Now()

		tokens := varnam.tokenizeWord(ctx, word, matchType, partial)

		if LOG_TIME_TAKEN {
			log.Printf("%s took %v\n", "channelTokenizeWord", time.Since(start))
		}

		channel <- tokens
		close(channel)
	}
}

func (varnam *Varnam) channelTokensToSuggestions(ctx context.Context, tokens *[]Token, limit int, channel chan []Suggestion) {
	select {
	case <-ctx.Done():
		close(channel)
		return
	default:
		start := time.Now()

		sugs := varnam.tokensToSuggestions(ctx, tokens, false, limit)

		if LOG_TIME_TAKEN {
			log.Printf("%s took %v\n", "channelTokensToSuggestions", time.Since(start))
		}

		channel <- sugs
		close(channel)
	}
}

func (varnam *Varnam) channelTokensToGreedySuggestions(ctx context.Context, tokens *[]Token, channel chan []Suggestion) {
	select {
	case <-ctx.Done():
		close(channel)
		return
	default:
		start := time.Now()

		sugs := varnam.tokensToSuggestions(ctx, tokens, false, varnam.TokenizerSuggestionsLimit)

		if LOG_TIME_TAKEN {
			log.Printf("%s took %v\n", "channelTokensToGreedySuggestions", time.Since(start))
		}

		channel <- sugs
		close(channel)
	}
}

func (varnam *Varnam) channelGetFromDictionary(ctx context.Context, word string, tokens *[]Token, channel chan channelDictionaryResult) {
	var (
		dictResults  []Suggestion
		exactMatches []Suggestion
	)

	select {
	case <-ctx.Done():
		close(channel)
		return
	default:
		start := time.Now()

		dictSugs := varnam.getFromDictionary(ctx, tokens)

		if varnam.Debug {
			fmt.Println("Dictionary results:", dictSugs)
		}

		if len(dictSugs.sugs) > 0 {
			if dictSugs.exactMatch == false {
				// These will be partial words
				restOfWord := word[dictSugs.longestMatchPosition+1:]

				start := time.Now()

				dictResults = varnam.tokenizeRestOfWord(ctx, restOfWord, dictSugs.sugs, varnam.DictionarySuggestionsLimit)

				if LOG_TIME_TAKEN {
					log.Printf("%s took %v\n", "tokenizeRestOfWord", time.Since(start))
				}
			} else {
				exactMatches = dictSugs.sugs

				start := time.Now()

				// Since partial words are in dictionary, exactMatch will be TRUE
				// for pathway to a word. Hence we're calling this here
				moreFromDict := varnam.getMoreFromDictionary(ctx, dictSugs.sugs)

				if varnam.Debug {
					fmt.Println("More dictionary results:", moreFromDict)
				}

				for _, sugSet := range moreFromDict {
					dictResults = append(dictResults, sugSet...)
				}

				if LOG_TIME_TAKEN {
					log.Printf("%s took %v\n", "getMoreFromDictionary", time.Since(start))
				}
			}
		}

		if LOG_TIME_TAKEN {
			log.Printf("%s took %v\n", "channelGetFromDictionary", time.Since(start))
		}

		channel <- channelDictionaryResult{exactMatches, dictResults}
		close(channel)
	}
}

func (varnam *Varnam) channelGetFromPatternDictionary(ctx context.Context, word string, channel chan channelDictionaryResult) {
	var (
		dictResults  []Suggestion
		exactMatches []Suggestion
	)

	select {
	case <-ctx.Done():
		close(channel)
		return
	default:
		start := time.Now()

		patternDictSugs := varnam.getFromPatternDictionary(ctx, word)

		if len(patternDictSugs) > 0 {
			if varnam.Debug {
				fmt.Println("Pattern dictionary results:", patternDictSugs)
			}

			var partialMatches []PatternDictionarySuggestion

			for _, match := range patternDictSugs {
				if match.Length < len(word) {
					sug := &match.Sug

					// Increase weight on length matched.
					// 50 because half of 100%
					sug.Weight += match.Length * 50

					for _, cb := range varnam.PatternWordPartializers {
						cb(sug)
					}

					partialMatches = append(partialMatches, match)
				} else if match.Length == len(word) {
					// Same length
					exactMatches = append(exactMatches, match.Sug)
				} else {
					dictResults = append(dictResults, match.Sug)
				}
			}

			perMatchLimit := varnam.PatternDictionarySuggestionsLimit

			if len(partialMatches) > 0 && perMatchLimit > len(partialMatches) {
				perMatchLimit = perMatchLimit / len(partialMatches)
			}

			for _, match := range partialMatches {
				restOfWord := word[match.Length:]

				filled := varnam.tokenizeRestOfWord(ctx, restOfWord, []Suggestion{match.Sug}, perMatchLimit)

				dictResults = append(dictResults, filled...)

				if len(dictResults) >= varnam.PatternDictionarySuggestionsLimit {
					break
				}
			}
		}

		if LOG_TIME_TAKEN {
			log.Printf("%s took %v\n", "channelGetFromPatternDictionary", time.Since(start))
		}

		channel <- channelDictionaryResult{exactMatches, dictResults}
		close(channel)
	}
}

func (varnam *Varnam) channelGetMoreFromDictionary(ctx context.Context, sugs []Suggestion, channel chan [][]Suggestion) {
	select {
	case <-ctx.Done():
		close(channel)
		return
	default:
		start := time.Now()

		result := varnam.getMoreFromDictionary(ctx, sugs)

		if LOG_TIME_TAKEN {
			log.Printf("%s took %v\n", "channelGetMoreFromDictionary", time.Since(start))
		}

		channel <- result
		close(channel)
	}
}
