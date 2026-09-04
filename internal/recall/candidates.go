package recall

type Candidates struct {
	Daily  []*Snapshot
	Hourly []*Snapshot
	Latest []*Snapshot
}

func (c *Candidates) Count() int {
	count := len(c.Latest) + len(c.Hourly) + len(c.Daily)
	return count
}

func (c *Candidates) GetSnapshotByAbsIndex(idx int) *Snapshot {
	if idx >= c.Count() {
		return nil
	}

	if idx < len(c.Daily) {
		return c.Daily[idx]
	}
	idx -= len(c.Daily)

	if idx < len(c.Hourly) {
		return c.Hourly[idx]
	}
	idx -= len(c.Hourly)

	if idx < len(c.Latest) {
		return c.Latest[idx]
	}
	idx -= len(c.Latest)

	return nil
}
