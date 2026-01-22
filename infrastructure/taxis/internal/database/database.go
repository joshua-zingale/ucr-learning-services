package database

func GetGroups(userId string) ([]string, error) {
	return []string{"cs100.instructor", "cs287.student"}, nil
}
