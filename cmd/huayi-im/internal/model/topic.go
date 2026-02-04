package model

import (
	"sync"
)

// Topic 话题模型
type Topic struct {
	Name      string   `json:"name"`
	Users     []string `json:"users"`
	CreatedAt string   `json:"created_at"`
}

// TopicManager Topic管理器
type TopicManager struct {
	topics map[string]*Topic
	mutex  sync.RWMutex
}

// NewTopicManager 创建Topic管理器实例
func NewTopicManager() *TopicManager {
	return &TopicManager{
		topics: make(map[string]*Topic),
	}
}

// CreateTopic 创建新Topic
func (tm *TopicManager) CreateTopic(name string) *Topic {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if topic, exists := tm.topics[name]; exists {
		return topic
	}

	topic := &Topic{
		Name:      name,
		Users:     []string{},
		CreatedAt: "",
	}
	tm.topics[name] = topic
	return topic
}

// GetTopic 获取Topic
func (tm *TopicManager) GetTopic(name string) (*Topic, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	topic, exists := tm.topics[name]
	return topic, exists
}

// DeleteTopic 删除Topic
func (tm *TopicManager) DeleteTopic(name string) bool {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if _, exists := tm.topics[name]; !exists {
		return false
	}

	delete(tm.topics, name)
	return true
}

// AddUserToTopic 添加用户到Topic
func (tm *TopicManager) AddUserToTopic(topicName, username string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	topic, exists := tm.topics[topicName]
	if !exists {
		// 如果Topic不存在，创建它
		topic = &Topic{
			Name:      topicName,
			Users:     []string{username},
			CreatedAt: "",
		}
		tm.topics[topicName] = topic
		return
	}

	// 检查用户是否已在Topic中
	for _, user := range topic.Users {
		if user == username {
			return
		}
	}

	// 添加用户到Topic
	topic.Users = append(topic.Users, username)
}

// RemoveUserFromTopic 从Topic中移除用户
func (tm *TopicManager) RemoveUserFromTopic(topicName, username string) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	topic, exists := tm.topics[topicName]
	if !exists {
		return
	}

	// 查找并移除用户
	for i, user := range topic.Users {
		if user == username {
			topic.Users = append(topic.Users[:i], topic.Users[i+1:]...)
			break
		}
	}

	// 如果Topic中没有用户了，删除Topic
	if len(topic.Users) == 0 {
		delete(tm.topics, topicName)
	}
}

// GetAllTopics 获取所有Topic
func (tm *TopicManager) GetAllTopics() []*Topic {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	var topics []*Topic
	for _, topic := range tm.topics {
		topics = append(topics, topic)
	}
	return topics
}

// GetTopicUsers 获取Topic中的用户
func (tm *TopicManager) GetTopicUsers(topicName string) ([]string, bool) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	topic, exists := tm.topics[topicName]
	if !exists {
		return nil, false
	}

	return topic.Users, true
}

// IsUserInTopic 检查用户是否在Topic中
func (tm *TopicManager) IsUserInTopic(topicName, username string) bool {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	topic, exists := tm.topics[topicName]
	if !exists {
		return false
	}

	for _, user := range topic.Users {
		if user == username {
			return true
		}
	}
	return false
}
