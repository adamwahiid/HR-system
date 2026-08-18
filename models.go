package main

type Role struct {
	RoleID   int    `gorm:"column:role_id;primaryKey;autoIncrement" json:"role_id"`
	RoleName string `gorm:"column:role_name;unique;not null" json:"role_name"`
}

func (Role) TableName() string {
	return "roles"
}

type User struct {
	UserID       int    `gorm:"column:user_id;primaryKey;autoIncrement" json:"user_id"`
	Email        string `gorm:"column:email;unique;not null" json:"email"`
	PasswordHash string `gorm:"column:password_hash;not null" json:"-"`
	RoleID       int    `gorm:"column:role_id;not null" json:"role_id"`
}

func (User) TableName() string {
	return "users"
}

type Worker struct {
	WorkerID  int     `gorm:"column:worker_id;primaryKey;autoIncrement" json:"worker_id"`
	FirstName string  `gorm:"column:first_name;not null" json:"first_name"`
	LastName  string  `gorm:"column:last_name;not null" json:"last_name"`
	Email     string  `gorm:"column:email;unique;not null" json:"email"`
	Password  string  `gorm:"column:password;not null" json:"password"`
	Salary    float64 `gorm:"column:salary;not null" json:"salary"`
	RoleID    int     `gorm:"column:role_id;not null" json:"role_id"`
	ManagerID int     `gorm:"column:manager_id;not null" json:"manager_id"`
}

func (Worker) TableName() string {
	return "workers"
}

type Manager struct {
	ManagerID int     `gorm:"column:manager_id;primaryKey;autoIncrement" json:"manager_id"`
	FirstName string  `gorm:"column:first_name;not null" json:"first_name"`
	LastName  string  `gorm:"column:last_name;not null" json:"last_name"`
	Email     string  `gorm:"column:email;unique;not null" json:"email"`
	Password  string  `gorm:"column:password;not null" json:"password"`
	Salary    float64 `gorm:"column:salary;not null" json:"salary"`
	RoleID    int     `gorm:"column:role_id;not null" json:"role_id"`
	BoardMemID int     `gorm:"column:board_mem_id" json:"board_mem_id"`
}

func (Manager) TableName() string {
	return "managers"
}

type BoardMember struct {
	BoardMemID int     `gorm:"column:board_mem_id;primaryKey;autoIncrement" json:"board_mem_id"`
	FirstName  string  `gorm:"column:first_name;not null" json:"first_name"`
	LastName   string  `gorm:"column:last_name;not null" json:"last_name"`
	Email      string  `gorm:"column:email;unique;not null" json:"email"`
	Password   string  `gorm:"column:password;not null" json:"password"`
	Salary     float64 `gorm:"column:salary;not null" json:"salary"`
	RoleID     int     `gorm:"column:role_id;not null" json:"role_id"`
}

func (BoardMember) TableName() string {
	return "board_members"
}
