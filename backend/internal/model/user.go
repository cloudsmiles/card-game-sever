package model

import (
	"time"

	"gorm.io/gorm"
)

// User 用户模型
type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	OpenID    string         `gorm:"size:128;uniqueIndex" json:"open_id,omitempty"` // 微信 openid
	UnionID   string         `gorm:"size:128;index" json:"union_id,omitempty"`      // 微信 unionid
	Nickname  string         `gorm:"size:64;not null" json:"nickname"`
	AvatarURL string         `gorm:"size:512" json:"avatar_url,omitempty"`
	LoginType string         `gorm:"size:20;not null;index" json:"login_type"` // wechat_h5 | wechat_open | guest
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

// UserScore 用户游戏积分
type UserScore struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"not null;uniqueIndex:idx_user_game" json:"user_id"`
	GameType    string    `gorm:"size:20;not null;uniqueIndex:idx_user_game" json:"game_type"`
	Score       int       `gorm:"not null;default:0" json:"score"`
	GamesPlayed int       `gorm:"not null;default:0" json:"games_played"`
	GamesWon    int       `gorm:"not null;default:0" json:"games_won"`
	UpdatedAt   time.Time `json:"updated_at"`

	User User `gorm:"foreignKey:UserID" json:"-"`
}

// FindUserByOpenID 根据微信 openid 查找用户
func FindUserByOpenID(openID string) (*User, error) {
	var user User
	err := DB.Where("open_id = ?", openID).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindUserByID 根据 ID 查找用户
func FindUserByID(id uint) (*User, error) {
	var user User
	err := DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// CreateUser 创建用户
func CreateUser(user *User) error {
	return DB.Create(user).Error
}

// UpdateUser 更新用户信息
func UpdateUser(user *User) error {
	return DB.Save(user).Error
}

// FindOrCreateWeChatUser 查找或创建微信用户
func FindOrCreateWeChatUser(openID, unionID, nickname, avatarURL, loginType string) (*User, error) {
	var user User
	err := DB.Where("open_id = ?", openID).First(&user).Error
	if err == nil {
		// 已存在，更新信息
		user.Nickname = nickname
		user.AvatarURL = avatarURL
		user.UnionID = unionID
		if err := DB.Save(&user).Error; err != nil {
			return nil, err
		}
		return &user, nil
	}
	// 不存在，创建新用户
	user = User{
		OpenID:    openID,
		UnionID:   unionID,
		Nickname:  nickname,
		AvatarURL: avatarURL,
		LoginType: loginType,
	}
	if err := DB.Create(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
