package kadnet

type ReplyBuffers struct {
	nodeReplyBuffer *NodeReplyBuffer
}

func (r *ReplyBuffers) GetNodeReplyBuffer() *NodeReplyBuffer {
	return r.nodeReplyBuffer
}
