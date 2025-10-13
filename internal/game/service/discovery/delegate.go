	package discovery


	//implementa a interface memberlist. Delegate é responsável
	// por fornecer informações locais (metadata) deste servidor para os demais
	// membros do cluster.	
	// ========================================
// Delegate (metadata do memberlist)
// ========================================

type delegate struct {
	meta []byte
}

func (d *delegate) NodeMeta(limit int) []byte {
	return d.meta
}

func (d *delegate) NotifyMsg([]byte) {}
func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}
func (d *delegate) LocalState(join bool) []byte {
	return nil
}
func (d *delegate) MergeRemoteState(buf []byte, join bool) {}
