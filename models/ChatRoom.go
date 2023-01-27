package models

import (
	"EasyChat/dbConn"
	"log"
	"math/rand"

	sf "EasyChat/utils/snowFlake"

	"github.com/garyburd/redigo/redis"
)

const (
	RoomMemberTable       = "EasyChat:RoomMember"
	RoomMemberTableExpire = 600
)

func init() {
	redisPool = dbConn.NewRedisPool()
}

var defaultLetters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

type ChatRoom struct {
	RoomKey  string `gorm:"column:room_key" json:"room_key"`
	MemberId int64  `gorm:"column:member_id" json:"member_id"`
}

func (ChatRoom) TableName() string {
	return "chat_room"
}

// 创建聊天室Key
func NewRoomKey() string {
	b := make([]rune, 10)
	for i := range b {
		//使用雪花算法生成的id作为随机数种子
		rand.Seed(sf.GenID())
		b[i] = defaultLetters[rand.Intn(len(defaultLetters))]
	}
	return string(b)
}

// 查询该聊天室中是否存在该成员
func (chatRoom *ChatRoom) CheckMember(RoomKey string, MemberId int64) (bool, error) {
	var count int64
	result := DB.Model(&ChatRoom{}).Where(&ChatRoom{RoomKey: RoomKey, MemberId: MemberId}).Count(&count)
	if result.Error != nil {
		return true, result.Error
	}
	//该聊天室已存在该成员
	if count != 0 {
		return true, nil
	}
	//该聊天室不存在该成员
	return false, nil
}

// 向聊天室中添加成员
func (chatRoom *ChatRoom) AddMember(RoomKey string, MemberId int64) error {

	//先在Mysql中创建记录
	chatRoom.RoomKey = RoomKey
	chatRoom.MemberId = MemberId
	result := DB.Create(&chatRoom)
	if result.Error != nil {
		return result.Error
	}

	//然后在redis中创建记录，使用列表LIST
	rc := redisPool.Get()
	defer rc.Close()
	_, err := rc.Do("LPUSH", RoomMemberTable+":"+RoomKey, MemberId)
	if err != nil {
		return err
	}
	//设置List的过期时间
	_, err1 := rc.Do("EXPIRE", RoomMemberTable+":"+RoomKey, RoomMemberTableExpire)
	if err1 != nil {
		return err1
	}

	return nil

}

// 获取聊天室所有成员的id
func (chatRoom *ChatRoom) GetMembers(room_key string) ([]int64, error) {

	//（1）先从缓存(redis)中取
	rc := redisPool.Get()
	defer rc.Close()
	//获取列表的长度
	listLen, err := redis.Int(rc.Do("LLEN", RoomMemberTable+":"+room_key))
	if err != nil {
		log.Println(err)
	}
	//获取列表的所有内容
	member_ids, err := redis.Int64s(rc.Do("LRANGE", RoomMemberTable+":"+room_key, 0, listLen-1))
	if err != nil {
		log.Println(err)
		log.Println("cache missing")
	}
	if len(member_ids) != 0 {
		log.Println("cache hits")
	}

	//（2）若缓存不存在，则从Mysql中取
	if len(member_ids) == 0 {
		chatRooms := []ChatRoom{}
		result := DB.Select("member_id").Where("room_key = ?", room_key).Find(&chatRooms)
		if result.Error != nil {
			return nil, err
		}
		//同时重新设置缓存
		for _, chatRoom := range chatRooms {
			//加入列表
			_, err := rc.Do("LPUSH", RoomMemberTable+":"+room_key, chatRoom.MemberId)
			if err != nil {
				return nil, err
			}
			//加入member_ids
			member_ids = append(member_ids, chatRoom.MemberId)
		}
		//设置整个List的过期时间
		_, err1 := rc.Do("EXPIRE", RoomMemberTable+":"+room_key, RoomMemberTableExpire)
		if err1 != nil {
			return nil, err1
		}
	}

	return member_ids, nil

}
