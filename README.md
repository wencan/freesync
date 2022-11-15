# freesync

[![Go Reference](https://pkg.go.dev/badge/github.com/wencan/freesync)](https://pkg.go.dev/github.com/wencan/freesync)  

Concurrency safety data structures and algorithms based on lock-free.

本项目包含两个部分，freesyn/lockfree为一套无锁的基础数据结构。freesync为一套基于无锁基础数据结构的简单复合结构。

## 目录
| 包 | 结构 | 说明 | 性能 |
| -- | --- | --- | ---- |
| freesync/lockfree | LimitedSlice | 无锁的长度受限的Slice | |
| freesync/lockfree | SinglyLinkedList | 无锁的单链表 | |
| freesync/lockfree | Slice | 无锁的支持增长的Slice | |
| freesync | Slice | 并发安全的Slice | 	与官方slice+mutex相比，写性能提升一半，读性能提升百倍左右 |
| freesync | Bag | 并发安全的容器 | 与sync.Map相比，写性能提升一半左右 |