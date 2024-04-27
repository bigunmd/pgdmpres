package util

import (
	"strings"

	"github.com/rs/zerolog/log"

	"github.com/common-nighthawk/go-figure"
)

func PrintBanner(s string) {
	for _, s := range figure.NewFigure(strings.ToUpper(s), "lcd", true).Slicify() {
		log.Info().Msg(s)
	}
}