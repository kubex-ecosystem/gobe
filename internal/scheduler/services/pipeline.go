package services

type Pipeline struct {
	buffer int
}

func (p *Pipeline) Run(data []int) []int {
	in := make(chan int, p.buffer)
	out := make(chan int, p.buffer)

	// Stage 1: Feed data
	go func() {
		for _, d := range data {
			in <- d
		}
		close(in)
	}()

	// Stage 2: Process
	go func() {
		for d := range in {
			out <- d * 2
		}
		close(out)
	}()

	// Collect
	var results []int
	for r := range out {
		results = append(results, r)
	}
	return results
}
