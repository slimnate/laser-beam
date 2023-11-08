package organization

type Organization struct {
	ID   int64
	Name string
}

type OrganizationSecret struct {
	Organization
	Key string
}
