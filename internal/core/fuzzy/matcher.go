package fuzzy

// Placeholder for fuzzy matching logic

// LevenshteinDistance computes the Levenshtein distance between two strings
func LevenshteinDistance(a, b string) int {
	la := len(a)
	lb := len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}
	dp := make([][]int, la+1)
	for i := range dp {
		dp[i] = make([]int, lb+1)
	}
	for i := 0; i <= la; i++ {
		dp[i][0] = i
	}
	for j := 0; j <= lb; j++ {
		dp[0][j] = j
	}
	for i := 1; i <= la; i++ {
		for j := 1; j <= lb; j++ {
			cost := 0
			if a[i-1] != b[j-1] {
				cost = 1
			}
			dp[i][j] = min(
				dp[i-1][j]+1,      // deletion
				dp[i][j-1]+1,      // insertion
				dp[i-1][j-1]+cost, // substitution
			)
		}
	}
	return dp[la][lb]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

// FuzzyMatchToolName returns the closest tool name and its distance, or "" if no match is close enough
func FuzzyMatchToolName(input string, known []string, maxDistance int) (string, int) {
	closest := ""
	minDist := maxDistance + 1
	for _, name := range known {
		dist := LevenshteinDistance(input, name)
		if dist < minDist {
			closest = name
			minDist = dist
		}
	}
	if minDist <= maxDistance {
		return closest, minDist
	}
	return "", -1
}
