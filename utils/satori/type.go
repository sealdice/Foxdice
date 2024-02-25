package satori

type ChannelType int

const (
	TextChannel ChannelType = iota
	DirectChannel
	CategoryChannel
	VoiceChannel
)

type Channel struct {
	Id       string
	Type     ChannelType
	Name     string
	ParentId string
}

type Guild struct {
	Id     string
	Name   string
	Avatar string
}

type GuildMember struct {
	User     User
	Nick     string
	Avatar   string
	JoinedAt int64
}

type GuildRole struct {
	Id   string
	Name string
}

type User struct {
	Id     string
	Name   string
	Nick   string
	Avatar string
	IsBot  bool
}

type Message struct {
	Id        string
	Context   string
	Channel   Channel
	Guild     Guild
	Member    GuildMember
	User      User
	CreatedAt int64
	UpdatedAt int64
}

type Status int

const (
	Offline Status = iota
	OnLine
	Connect
	DisConnect
	ReConnect
)

type Login struct {
	User     User
	SelfId   string
	Platform string
	Status   Status
}

type Argv struct {
	Name      string
	Arguments []any
	Option    any
}

type Button struct {
	Id string
}

type Event struct {
	Id        int64
	Type      string
	Platform  string
	SelfId    string
	Timestamp int
	Argv      Argv
	Button    Button
	Channel   Channel
	Guild     Guild
	Login     Login
	Member    GuildMember
	Message   Message
	Operator  User
	Role      GuildRole
	User      User
}
