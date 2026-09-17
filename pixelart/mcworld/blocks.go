package mcworld

import (
	"image/color"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/Sanjar0126/go-pixel-art/pixelart"
)

// BlockColor is a placeable Minecraft block id together with a
// representative average color sampled from its texture.
type BlockColor struct {
	ID    string
	Color color.Color
}

// nonBlockKeywords filters out icon textures that do not correspond to a
// placeable block: GUI/HUD art, item/entity icons, effects, etc. This is a
// heuristic; tune it by editing this list if the wrong textures slip in or
// out.
var nonBlockKeywords = []string{
	"gui_", "icon", "_icon", "spawn_egg", "particle", "effect_", "entity_",
	"banner", "shield", "trim_", "template", "empty", "debug", "widget",
	"button", "slot", "background", "overlay", "destroy_stage", "breaking",
	"painting", "book", "potion", "arrow", "trident", "bow_", "fishing_rod",
	"boat", "minecart", "map_", "compass", "clock", "saddle", "horse",
	"armor_", "elytra", "totem", "shulker_box_", "sign", "_sign",
	"bed_", "banner_", "villager", "mob_", "player_", "skull", "head",
	"chest_boat", "music_disc", "leather_", "trident_", "crossbow",
}

// stripSuffixes removes texture-part suffixes (side/top/etc.) so multiple
// texture files for one block collapse to a single block id where possible.
var stripSuffixRe = regexp.MustCompile(`_(top|bottom|side|front|back|inner|outer|stage[0-9]+|on|lit|open|closed|particle)$`)

// BuildPaletteFromIcons scans dir for PNG textures, filters out anything
// that looks like a non-block icon, and returns one BlockColor per
// remaining texture using its average color. The block id is derived from
// the filename as "minecraft:<name>".
func BuildPaletteFromIcons(dir string) ([]BlockColor, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var out []BlockColor
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasSuffix(strings.ToLower(name), ".png") {
			continue
		}
		base := strings.TrimSuffix(name, filepath.Ext(name))
		lower := strings.ToLower(base)
		if isLikelyNonBlock(lower) {
			continue
		}
		blockID := "minecraft:" + stripSuffixRe.ReplaceAllString(lower, "")
		if seen[blockID] {
			continue
		}

		img, err := pixelart.LoadImage(filepath.Join(dir, name))
		if err != nil {
			continue
		}
		seen[blockID] = true
		out = append(out, BlockColor{ID: blockID, Color: pixelart.AverageColor(img)})
	}

	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func isLikelyNonBlock(lowerName string) bool {
	for _, kw := range nonBlockKeywords {
		if strings.Contains(lowerName, kw) {
			return true
		}
	}
	return false
}

// closestBlock returns the palette entry whose color is nearest to c.
func closestBlock(c color.Color, palette []BlockColor) BlockColor {
	r1, g1, b1, _ := c.RGBA()
	best := palette[0]
	bestDist := int64(1) << 62
	for _, bc := range palette {
		r2, g2, b2, _ := bc.Color.RGBA()
		dr := int64(r1>>8) - int64(r2>>8)
		dg := int64(g1>>8) - int64(g2>>8)
		db := int64(b1>>8) - int64(b2>>8)
		dist := dr*dr + dg*dg + db*db
		if dist < bestDist {
			bestDist = dist
			best = bc
		}
	}
	return best
}
