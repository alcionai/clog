package clog

import "github.com/alcionai/clues"

func setCluesSecretsHash(alg piiAlg) {
	switch alg {
	case PIIHash:
		clues.SetHasher(clues.DefaultHash())
	case PIIMask:
		clues.SetHasher(clues.HashCfg{HashAlg: clues.Flatmask})
	case PIIPlainText:
		clues.SetHasher(clues.NoHash())
	}
}
