package mpts

// Option are options for the MultiPointToSinglePointCompilation
type Option func(*MultipointToSinglepointCompilation)

// WithNumRowProfileOpt tells the compiler to add shadow columns (columns
// equal to zero everywhere to the comp). These columns are added so that
// the number of columns for each round matches the sizes indicated in the
// profile. The positions in the provided slices are understood as "starting
// from the first non-empty" rounds that are compiled by the current
// compilation context.
func WithNumRowProfileOpt(numRowProfileOpt []int) Option {
	return func(ctx *MultipointToSinglepointCompilation) {
		ctx.NumColumnProfileOpt = numRowProfileOpt
	}
}
