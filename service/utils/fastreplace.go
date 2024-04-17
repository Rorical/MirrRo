package utils

type State struct {
	transitions map[rune]*State
	isFinal     bool
}

func NewState() *State {
	return &State{transitions: make(map[rune]*State)}
}

type DFA struct {
	root           *State
	replacementMap map[string]string
}

func NewDFA() *DFA {
	return &DFA{
		root:           NewState(),
		replacementMap: nil,
	}
}

func (d *DFA) AddWord(word string) {
	current := d.root
	for _, ch := range word {
		if next, ok := current.transitions[ch]; ok {
			current = next
		} else {
			newState := NewState()
			current.transitions[ch] = newState
			current = newState
		}
	}
	current.isFinal = true
}

func (d *DFA) Build(replacementMap map[string]string) {
	for word := range replacementMap {
		d.AddWord(word)
	}
	d.replacementMap = replacementMap
}

func (d *DFA) ReplaceAll(text string) string {
	result := ""
	current := d.root
	matchStart := -1
	lastIndex := 0

	for i, ch := range text {
		if next, ok := current.transitions[ch]; ok { // Transition exists.
			if matchStart == -1 { // Start of a new match.
				matchStart = i
			}
			current = next
			if current.isFinal { // End of a valid word.
				wordToReplace := text[matchStart : i+1]
				replacement, exists := d.replacementMap[wordToReplace]
				if exists {
					result += text[lastIndex:matchStart] + replacement
					lastIndex = i + 1
				}
				matchStart = -1  // Reset for potential next match.
				current = d.root // Reset DFA to root state.
			}
		} else {
			if matchStart != -1 { // No transition and we are in a potential match.
				i -= i - matchStart // Move back to the start of the failed match and continue.
				matchStart = -1
				current = d.root
			}
		}
	}

	result += text[lastIndex:] // Append remaining part of the text if any.

	return result
}
