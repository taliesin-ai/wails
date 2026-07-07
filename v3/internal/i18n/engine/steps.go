package engine

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	stepsOpenRe        = regexp.MustCompile(`^<Steps(?:\s|>|$)`)
	stepsCloseRe       = regexp.MustCompile(`^</Steps\s*>$`)
	stepsOrderedItemRe = regexp.MustCompile(`^\s*\d+\.\s+`)
)

func validateStepsStructure(srcBody, outBody string) []string {
	srcBlocks := stepsBlocks(srcBody)
	outBlocks := stepsBlocks(outBody)
	if len(srcBlocks) == 0 && len(outBlocks) == 0 {
		return nil
	}

	var warnings []string
	if len(srcBlocks) != len(outBlocks) {
		warnings = append(warnings, fmt.Sprintf("Steps block count changed (%d -> %d)", len(srcBlocks), len(outBlocks)))
	}
	for i := 0; i < min(len(srcBlocks), len(outBlocks)); i++ {
		if a, b := countStepsOrderedItems(srcBlocks[i]), countStepsOrderedItems(outBlocks[i]); a != b {
			warnings = append(warnings, fmt.Sprintf("Steps ordered-list item count changed in block %d (%d -> %d)", i+1, a, b))
		}
		if hasIndentedStepsContinuation(srcBlocks[i]) && hasFlushStepsContinuation(outBlocks[i]) {
			warnings = append(warnings, fmt.Sprintf("Steps ordered-list indentation changed in block %d", i+1))
		}
	}
	return warnings
}

func stepsBlocks(body string) []string {
	var blocks []string
	var block []string
	inSteps := false
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if !inSteps && stepsOpenRe.MatchString(trimmed) {
			inSteps = true
			block = nil
			continue
		}
		if !inSteps {
			continue
		}
		if stepsCloseRe.MatchString(trimmed) {
			blocks = append(blocks, strings.Join(block, "\n"))
			inSteps = false
			continue
		}
		block = append(block, line)
	}
	return blocks
}

func countStepsOrderedItems(block string) int {
	count := 0
	for _, line := range strings.Split(block, "\n") {
		if stepsOrderedItemRe.MatchString(line) {
			count++
		}
	}
	return count
}

func hasIndentedStepsContinuation(block string) bool {
	seenItem := false
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if stepsOrderedItemRe.MatchString(line) {
			seenItem = true
			continue
		}
		if seenItem && hasStepsContinuationIndent(line) {
			return true
		}
	}
	return false
}

func hasFlushStepsContinuation(block string) bool {
	seenItem := false
	for _, line := range strings.Split(block, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if stepsOrderedItemRe.MatchString(line) {
			seenItem = true
			continue
		}
		if seenItem && !hasStepsContinuationIndent(line) {
			return true
		}
	}
	return false
}

func hasStepsContinuationIndent(line string) bool {
	return strings.HasPrefix(line, "\t") || strings.HasPrefix(line, "   ")
}
