package diag

// dfsChannels is the set of channels subject to Dynamic Frequency Selection.
var dfsChannels = map[int]bool{
	52: true, 56: true, 60: true, 64: true,
	100: true, 104: true, 108: true, 112: true,
	116: true, 120: true, 124: true, 128: true,
	132: true, 136: true, 140: true, 144: true,
}

// IsDFSChannel returns true if the given channel is a DFS channel.
func IsDFSChannel(channel int) bool {
	return dfsChannels[channel]
}
