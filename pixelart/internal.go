package pixelart

import (
	"bufio"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"
	"time"
)

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func rgbToVec(c color.Color) ColorVec {
	r, g, b, _ := c.RGBA()
	return ColorVec{float64(r >> 8), float64(g >> 8), float64(b >> 8)}
}

func dist2(a, b ColorVec) float64 {
	dx := a[0] - b[0]
	dy := a[1] - b[1]
	dz := a[2] - b[2]
	return dx*dx + dy*dy + dz*dz
}

// --- Palette building (kmeans etc.) ---
// (same functions: kmeans, loadImage, sampleColorsFromImage, buildPaletteFromDir, nearestPaletteColor)

// kmeans runs a simple k-means on colors. Returns centroids.
func kmeans(samples []ColorVec, k int, maxIter int) []ColorVec {
	n := len(samples)
	if n == 0 || k <= 0 {
		return nil
	}
	if k >= n {
		cent := make([]ColorVec, k)
		copy(cent, samples)
		return cent
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	centroids := make([]ColorVec, k)
	perm := rng.Perm(n)
	for i := 0; i < k; i++ {
		centroids[i] = samples[perm[i]]
	}

	labels := make([]int, n)
	counts := make([]int, k)
	for iter := 0; iter < maxIter; iter++ {
		changed := false

		for i, s := range samples {
			best := 0
			bestD := dist2(s, centroids[0])
			for j := 1; j < k; j++ {
				d := dist2(s, centroids[j])
				if d < bestD {
					bestD = d
					best = j
				}
			}
			if labels[i] != best {
				changed = true
				labels[i] = best
			}
		}
		if !changed && iter > 0 {
			break
		}

		// recompute centroids
		for j := 0; j < k; j++ {
			counts[j] = 0
			centroids[j] = ColorVec{0, 0, 0}
		}
		for i, s := range samples {
			l := labels[i]
			centroids[l][0] += s[0]
			centroids[l][1] += s[1]
			centroids[l][2] += s[2]
			counts[l]++
		}
		for j := 0; j < k; j++ {
			if counts[j] > 0 {
				centroids[j][0] /= float64(counts[j])
				centroids[j][1] /= float64(counts[j])
				centroids[j][2] /= float64(counts[j])
			} else {
				// empty cluster: reinitialize to random sample
				centroids[j] = samples[rng.Intn(n)]
			}
		}
	}
	return centroids
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func sampleColorsFromImage(img image.Image, maxSamples int) []ColorVec {
	b := img.Bounds()
	w := b.Dx()
	h := b.Dy()
	if maxSamples <= 0 {
		maxSamples = w * h
	}
	step := int(math.Ceil(math.Sqrt(float64(w*h) / float64(maxSamples))))
	if step < 1 {
		step = 1
	}
	out := make([]ColorVec, 0, maxSamples)
	for y := b.Min.Y; y < b.Max.Y; y += step {
		for x := b.Min.X; x < b.Max.X; x += step {
			c := img.At(x, y)
			out = append(out, rgbToVec(c))
			if len(out) >= maxSamples {
				return out
			}
		}
	}
	return out
}

// --- Processing ---
// (same function: processImageToPixelArt)

// --- Saving ---
func saveImageAsPNG(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	w := bufio.NewWriter(f)
	err = png.Encode(w, img)
	if err != nil {
		return err
	}
	return w.Flush()
}

func ensureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
