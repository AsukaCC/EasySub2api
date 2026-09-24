// Package modeltrace ports ModelTrace's MIT-licensed global fingerprint scorer.
// Source: xqy2006/ModelTrace, revision 55a2e4a55170423b484d701e9a82ab62b268c811.
package modeltrace

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"math"
	"math/rand/v2"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"unicode"
)

const Revision = "55a2e4a55170423b484d701e9a82ab62b268c811"
const QueryCount = 3
const dimension = 355

//go:embed unified_bank.json
var bankJSON []byte

type feature struct {
	Mean         []float64     `json:"feature_mean"`
	Scale        []float64     `json:"feature_scale"`
	Basis        [][]float64   `json:"nuisance_basis"`
	Centroids    [][]float64   `json:"centroids"`
	Environments [][][]float64 `json:"environment_centroids"`
	Weight       float64       `json:"weight"`
}

type candidate struct {
	ID         string `json:"id"`
	Name       string `json:"display_name"`
	Family     string `json:"family"`
	FamilyName string `json:"family_name"`
}

type bank struct {
	Models []candidate `json:"models"`
	Robust struct {
		Hellinger feature `json:"hellinger"`
		Ordered   feature `json:"ordered_blocks"`
	} `json:"robust"`
	Calibration map[string]struct {
		Beta float64 `json:"beta"`
	} `json:"calibration"`
}

var loadBank = sync.OnceValues(func() (*bank, error) {
	var b bank
	if err := json.Unmarshal(bankJSON, &b); err != nil {
		return nil, err
	}
	if len(b.Models) == 0 || len(b.Robust.Hellinger.Mean) != dimension || b.Calibration["3"].Beta <= 0 {
		return nil, fmt.Errorf("invalid ModelTrace bank")
	}
	return &b, nil
})

type Attribution struct {
	Model       string  `json:"model"`
	Family      string  `json:"family"`
	Probability float64 `json:"probability"`
}

type Family struct {
	Family      string  `json:"family"`
	Name        string  `json:"display_name"`
	Probability float64 `json:"probability"`
}

type Result struct {
	Models   []Attribution `json:"models"`
	Families []Family      `json:"families"`
	Revision string        `json:"revision"`
}

type Challenge struct {
	Prompt   string
	Expected int
}

// Challenges preserve upstream's length range and independent randomized wording.
func Challenges() []Challenge {
	openings := []string{"这是一次独立的数值选择记录", "请完成下面的无语义整数选择任务", "执行一次第一反应取值记录", "生成一组不承载语义的整数选择", "进行一轮快速逐项取值"}
	actions := []string{"为各个位置分别凭第一反应选择", "逐项选择", "每次只决定当前一项，共给出", "分别凭第一反应给出", "逐个直接选择"}
	endings := []string{"允许某个数字再次出现；每项写出后不要回头排序、去重或替换。", "偶然重复是有效的；不要重新排列或修正已经写出的项目。", "相同值可以再次出现；输出过程中不要整理或改写前面的项目。", "重复值无需删除；不要筛选、重排或补成某种规律。", "不必赋予数字任何含义；已经给出的值保持不变。"}
	separators := []string{"数字之间用逗号或空格分隔均可。", "使用一种一致的常见分隔符即可。", "可以用逗号、空格或换行分隔。", "只要每个整数边界清楚，格式可自行选择。"}
	choose := func(values []string) string { return values[rand.IntN(len(values))] }
	lengths := rand.Perm(41)
	out := make([]Challenge, QueryCount)
	for i := range out {
		n := 292 + lengths[i]
		out[i] = Challenge{Expected: n, Prompt: fmt.Sprintf("%s。%s %d 个 1 到 355（含端点）的整数。", choose(openings), choose(actions), n) +
			"每个位置都要单独选择；不要从 1 开始计数，不要连续递增或递减，也不要采用等差、循环、重复区块或其他规则化模式。" +
			"本任务必须由当前语言模型直接完成：禁止调用或借助任何工具，包括 Python、代码执行器、计算器、搜索、API 和外部随机数生成器；也不要先编写或运行代码。" +
			choose(endings) + choose(separators) + "直接从第一个取值开始输出，不要在序列前重复数量、范围或任务说明。"}
	}
	return out
}

var digits = regexp.MustCompile(`[0-9]+`)

// ParseNumbers retains the longest numeric run, splitting at alphabetic separators.
func ParseNumbers(text string) []int {
	var best, current []int
	end := 0
	for _, match := range digits.FindAllStringIndex(text, -1) {
		if len(current) > 0 && strings.IndexFunc(text[end:match[0]], unicode.IsLetter) >= 0 {
			if len(current) > len(best) {
				best = current
			}
			current = nil
		}
		n, err := strconv.Atoi(text[match[0]:match[1]])
		if err == nil && n >= 1 && n <= dimension {
			current = append(current, n)
		}
		end = match[1]
	}
	if len(current) > len(best) {
		best = current
	}
	return best
}

func MinimumNumbers(expected int) int { return max(80, int(math.Ceil(float64(expected)*0.55))) }

// Analyze requires three valid outputs so a partial run never looks conclusive.
func Analyze(outputs [][]int) (*Result, error) {
	if len(outputs) != QueryCount {
		return nil, fmt.Errorf("three valid outputs required")
	}
	b, err := loadBank()
	if err != nil {
		return nil, err
	}
	scores := make([]float64, len(b.Models))
	for _, numbers := range outputs {
		if len(numbers) < 80 {
			return nil, fmt.Errorf("insufficient numbers")
		}
		counts := make([]float64, dimension)
		for _, n := range numbers {
			if n < 1 || n > dimension {
				return nil, fmt.Errorf("number out of range")
			}
			counts[n-1]++
		}
		for i := range counts {
			counts[i] = math.Sqrt((counts[i] + 0.5) / (float64(len(numbers)) + 0.5*dimension))
		}
		marginal := standardize(similarities(normalized(project(transform(counts, b.Robust.Hellinger), b.Robust.Hellinger.Basis)), b.Robust.Hellinger.Centroids))
		ordered := orderedScores(numbers, b.Robust.Ordered)
		w := b.Robust.Ordered.Weight
		for i := range scores {
			scores[i] += ((1-w)*marginal[i] + w*ordered[i]) / QueryCount
		}
	}
	beta := b.Calibration["3"].Beta
	maximum := scores[0]
	for _, score := range scores {
		maximum = math.Max(maximum, score)
	}
	total := 0.0
	for i := range scores {
		scores[i] = math.Exp(beta * (scores[i] - maximum))
		total += scores[i]
	}
	result := &Result{Revision: Revision}
	families := make(map[string]int)
	for i, model := range b.Models {
		p := scores[i] / total
		result.Models = append(result.Models, Attribution{Model: model.ID, Family: model.Family, Probability: p})
		index, ok := families[model.Family]
		if !ok {
			index = len(result.Families)
			families[model.Family] = index
			result.Families = append(result.Families, Family{Family: model.Family, Name: model.FamilyName})
		}
		result.Families[index].Probability += p
	}
	sort.SliceStable(result.Models, func(i, j int) bool { return result.Models[i].Probability > result.Models[j].Probability })
	return result, nil
}

func dot(a, b []float64) float64 {
	sum := 0.0
	for i, x := range a {
		sum += x * b[i]
	}
	return sum
}
func normalized(a []float64) []float64 {
	scale := math.Max(math.Sqrt(dot(a, a)), 1e-12)
	for i := range a {
		a[i] /= scale
	}
	return a
}
func standardize(a []float64) []float64 {
	mean := 0.0
	for _, x := range a {
		mean += x
	}
	mean /= float64(len(a))
	variance := 0.0
	for _, x := range a {
		variance += (x - mean) * (x - mean)
	}
	scale := math.Max(math.Sqrt(variance/float64(len(a))), 1e-12)
	for i := range a {
		a[i] = (a[i] - mean) / scale
	}
	return a
}
func transform(a []float64, f feature) []float64 {
	for i := range a {
		a[i] = (a[i] - f.Mean[i]) / f.Scale[i]
	}
	return a
}
func project(a []float64, basis [][]float64) []float64 {
	for _, vector := range basis {
		p := dot(a, vector)
		for i := range a {
			a[i] -= p * vector[i]
		}
	}
	return a
}
func similarities(a []float64, centroids [][]float64) []float64 {
	out := make([]float64, len(centroids))
	for i, c := range centroids {
		out[i] = dot(a, c)
	}
	return out
}
func orderedScores(numbers []int, f feature) []float64 {
	values := make([]float64, 0, 74)
	appendBins := func(bins []float64) {
		total := 0.0
		for _, n := range bins {
			total += n
		}
		for _, n := range bins {
			values = append(values, math.Sqrt(n/total))
		}
	}
	start := 0
	for block := 0; block < 4; block++ {
		size := len(numbers) / 4
		if block < len(numbers)%4 {
			size++
		}
		bins := make([]float64, 16)
		for i := range bins {
			bins[i] = 0.5
		}
		for _, n := range numbers[start : start+size] {
			bins[min(15, (n-1)*16/dimension)]++
		}
		appendBins(bins)
		start += size
	}
	last := make([]float64, 10)
	for i := range last {
		last[i] = 0.5
	}
	for _, n := range numbers {
		last[n%10]++
	}
	appendBins(last)
	values = transform(values, f)
	unit := normalized(append([]float64(nil), values...))
	template := make([]float64, len(f.Centroids))
	for i := range template {
		template[i] = math.Inf(-1)
	}
	for _, environment := range f.Environments {
		for i, c := range environment {
			template[i] = math.Max(template[i], dot(unit, c))
		}
	}
	template = standardize(template)
	nuisance := standardize(similarities(normalized(project(values, f.Basis)), f.Centroids))
	for i := range template {
		template[i] = 0.5*template[i] + 0.5*nuisance[i]
	}
	return standardize(template)
}
