package kvraft

type db interface {
	Get(key string) (string, Err)
	Put(key, value string) Err
	Append(key, value string) Err
}

type DbKV struct {
	KV map[string]string
}

func NewDbKV() *DbKV {
	return &DbKV{make(map[string]string)}
}

func (kv *DbKV) Get(key string) (string, Err) {
	if value, ok := kv.KV[key]; ok {
		return value, OK
	}
	return "", ErrNoKey
}

func (kv *DbKV) Put(key, value string) Err {
	kv.KV[key] = value
	return OK
}

func (kv *DbKV) Append(key, value string) Err {
	kv.KV[key] += value
	return OK
}
