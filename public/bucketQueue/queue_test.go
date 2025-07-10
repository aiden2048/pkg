package bucketQueue

import (
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestLoopTasks(t *testing.T) {
	for i := 0; i < 20000000; i++ {
		for ii := 0; ii < 5; ii++ {

		}
	}
	fmt.Println("ok")
}

func TestConcurQueue(t *testing.T) {
	q := NewBucketQueueWithBucketConcur("testQ", 10, 2)
	go q.Run()

	for i := 0; i < 1000; i++ {
		func(i int) {
			fmt.Println("task ", i, " ", time.Now().Unix())
			q.Reg(NewTask(int64(i), "taskA", func(task *Task) error {
				fmt.Println("i am task A, bucketID ", i)
				fmt.Println("finish task ", i, " ", time.Now().Unix())
				return nil
			}, 0, 2))
		}(i)
	}
	select {}
}

func TestBucketConcurQueue(t *testing.T) {
	q := NewBucketQueueWithBucketConcur("testQ", 10, 2)
	q.SetSize(2)
	go q.Run()

	q.Reg(NewTask(1, "taskA", func(task *Task) error {
		fmt.Println("i am task A, bucketID 1, delay 5")
		return nil
	}, 5, 2))

	q.Reg(NewTask(1, "taskB", func(task *Task) error {
		fmt.Println("i am task B, bucketID 1, delay 0, concur 1")
		return nil
	}, 0, 2))

	err := q.Reg(NewTask(1, "taskB", func(task *Task) error {
		fmt.Println("i am task B, bucketID 1, delay 0, concur 2")
		return nil
	}, 0, 2))
	if err != nil {
		t.Fatal(err)
		return
	}

	q.Reg(NewTask(1, "taskB", func(task *Task) error {
		fmt.Println("i am task B, bucketID 1, delay 0, concur 3")
		return nil
	}, 0, 2))

	q.Reg(NewTask(1, "taskC", func(task *Task) error {
		fmt.Println("i am task C, bucketID 2, delay 1")
		return nil
	}, 1, 2))

	q.Reg(NewTask(2, "taskA", func(task *Task) error {
		fmt.Println("i am task A,  bucketID 2, delay 5")
		return nil
	}, 5, 2))

	q.Reg(NewTask(2, "taskB", func(task *Task) error {
		fmt.Println("i am task B, bucketID 2, delay 0")
		return nil
	}, 0, 2))

	q.Reg(NewTask(2, "taskC", func(task *Task) error {
		fmt.Println("i am task C, bucketID 2, delay 1")
		return nil
	}, 1, 2))

	q.Reg(NewTask(3, "taskA", func(task *Task) error {
		fmt.Println("i am task A,  bucketID 3, delay 5")
		if task.GetAttempted() > 1 {
			return nil
		}
		return errors.New("retry error")
	}, 5, 2))

	q.Reg(NewTask(3, "taskB", func(task *Task) error {
		fmt.Println("i am task B, bucketID 3, delay 0")
		return nil
	}, 0, 2))

	q.Reg(NewTask(3, "taskC", func(task *Task) error {
		fmt.Println("i am task C, bucketID 3, delay 1")
		return nil
	}, 1, 2))

	select {}
}

func TestQueue(t *testing.T) {
	q := NewBucketQueue("testQ", 10)

	go q.Run()

	q.Reg(NewTask(1, "taskA", func(task *Task) error {
		fmt.Println("i am task A, bucketID 2, delay 5")
		return nil
	}, 5, 2))

	q.Reg(NewTask(1, "taskB", func(task *Task) error {
		fmt.Println("i am task B, bucketID 2, delay 0")
		return nil
	}, 0, 2))

	q.Reg(NewTask(1, "taskC", func(task *Task) error {
		fmt.Println("i am task C, bucketID 2, delay 1")
		return nil
	}, 1, 2))

	q.Reg(NewTask(2, "taskA", func(task *Task) error {
		fmt.Println("i am task A,  bucketID 2, delay 5")
		return nil
	}, 5, 2))

	q.Reg(NewTask(2, "taskB", func(task *Task) error {
		fmt.Println("i am task B, bucketID 2, delay 0")
		return nil
	}, 0, 2))

	q.Reg(NewTask(2, "taskC", func(task *Task) error {
		fmt.Println("i am task C, bucketID 2, delay 1")
		return nil
	}, 1, 2))

	q.Reg(NewTask(3, "taskA", func(task *Task) error {
		fmt.Println("i am task A,  bucketID 3, delay 5")
		if task.GetAttempted() > 1 {
			return nil
		}
		return errors.New("retry error")
	}, 5, 2))

	q.Reg(NewTask(3, "taskB", func(task *Task) error {
		fmt.Println("i am task B, bucketID 3, delay 0")
		return nil
	}, 0, 2))

	q.Reg(NewTask(3, "taskC", func(task *Task) error {
		fmt.Println("i am task C, bucketID 3, delay 1")
		return nil
	}, 1, 2))

	time.Sleep(time.Second * 20)
}
