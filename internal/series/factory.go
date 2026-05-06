package series

import pub "github.com/kochemajaka/gofocas/series"

// For returns the Strategy for the given public series constant.
func For(s pub.Series) Strategy {
	switch s {
	case pub.S0i:
		return Series0i()
	case pub.S15:
		return Series15()
	case pub.S15i:
		return Series15i()
	case pub.S16:
		return Series16()
	case pub.S16i:
		return Series16i()
	case pub.S18i:
		return Series18i()
	case pub.S21:
		return Series21()
	case pub.S30i:
		return Series30i()
	case pub.S31i:
		return Series31i()
	case pub.S32i:
		return Series32i()
	default:
		return Default()
	}
}
