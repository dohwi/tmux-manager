package update

import (
	"fmt"
	"strconv"
	"strings"
)

func normalizeTag(s string) string {
	s = strings.TrimSpace(s)
	s = strings.TrimPrefix(s, "v")
	return s
}

func compareSemver(a, b string) (int, error) {
	if a == b {
		return 0, nil
	}

	am, ap := splitSemver(a)
	bm, bp := splitSemver(b)

	for i := 0; i < 3; i++ {
		if am[i] != bm[i] {
			if am[i] < bm[i] {
				return -1, nil
			}
			return 1, nil
		}
	}

	switch {
	case ap == "" && bp != "":
		return 1, nil
	case ap != "" && bp == "":
		return -1, nil
	case ap == "" && bp == "":
		return 0, nil
	}

	return comparePrerelease(ap, bp)
}

func splitSemver(s string) ([3]int, string) {
	var nums [3]int
	pre := ""

	if i := strings.Index(s, "+"); i >= 0 {
		s = s[:i]
	}
	if i := strings.Index(s, "-"); i >= 0 {
		pre = s[i+1:]
		s = s[:i]
	}

	parts := strings.SplitN(s, ".", 4)
	for i := 0; i < 3 && i < len(parts); i++ {
		n, err := strconv.Atoi(parts[i])
		if err == nil {
			nums[i] = n
		}
	}
	return nums, pre
}

func comparePrerelease(a, b string) (int, error) {
	if a == b {
		return 0, nil
	}
	at := strings.Split(a, ".")
	bt := strings.Split(b, ".")

	n := len(at)
	if len(bt) < n {
		n = len(bt)
	}
	for i := 0; i < n; i++ {
		ai, aIsNum := strconv.Atoi(at[i])
		bi, bIsNum := strconv.Atoi(bt[i])
		switch {
		case aIsNum == nil && bIsNum == nil:
			if ai != bi {
				if ai < bi {
					return -1, nil
				}
				return 1, nil
			}
		case aIsNum == nil && bIsNum != nil:
			return -1, nil
		case aIsNum != nil && bIsNum == nil:
			return 1, nil
		default:
			if at[i] != bt[i] {
				if at[i] < bt[i] {
					return -1, nil
				}
				return 1, nil
			}
		}
	}
	if len(at) != len(bt) {
		if len(at) < len(bt) {
			return -1, nil
		}
		return 1, nil
	}
	return 0, fmt.Errorf("incomparable prereleases: %q vs %q", a, b)
}
