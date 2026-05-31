package messaging

type Context struct {
	Kind          Kind
	Name          string
	AggregateType string
	Action        string
	CorrelationID *string
	CausationID   *string

	IPAddress *string
	UserAgent *string
	DeviceID  *string
}

func (c Context) ToMetadata() map[string]string {
	meta := map[string]string{
		"message_kind":   string(c.Kind),
		"message_name":   c.Name,
		"aggregate_type": c.AggregateType,
		"action":         c.Action,
	}

	if c.CorrelationID != nil {
		meta["correlation_id"] = *c.CorrelationID
	}
	if c.CausationID != nil {
		meta["causation_id"] = *c.CausationID
	}
	if c.IPAddress != nil {
		meta["ip_address"] = *c.IPAddress
	}
	if c.UserAgent != nil {
		meta["user_agent"] = *c.UserAgent
	}
	if c.DeviceID != nil {
		meta["device_id"] = *c.DeviceID
	}

	return meta
}
