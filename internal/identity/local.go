package identity

type LocalIdentity struct {
	UserID      string
	DisplayName string
	Source      string
}

func LocalDevIdentity() LocalIdentity {
	return LocalIdentity{
		UserID:      "user_local_dev",
		DisplayName: "Local Developer",
		Source:      "local_dev",
	}
}

func ResolveLocalIdentity(_ string) LocalIdentity {
	// M3 故意忽略外部 user 选择，避免伪多用户语义污染后续鉴权边界。
	return LocalDevIdentity()
}
