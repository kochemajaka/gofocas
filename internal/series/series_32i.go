package series

type series32i struct{ defaultStrategy }

func Series32i() Strategy { return series32i{} }
