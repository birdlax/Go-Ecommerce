// repository/unit_of_work.go
package repository

import "gorm.io/gorm"

// Repositories คือ struct ที่รวบรวม Repository ทั้งหมดที่เรามี
// เพื่อให้ Service สามารถเรียกใช้ได้ภายใน Transaction เดียวกัน
type Repositories struct {
	Product ProductRepository
	// Category CategoryRepository // ในอนาคตถ้ามี CategoryRepo ก็จะมาอยู่ที่นี่
	// User     UserRepository     // ...
}

// UnitOfWork คือ Interface ที่ Service จะเรียกใช้
// มีหน้าที่หลักคือการ Execute งานที่ต้องการ Transaction
type UnitOfWork interface {
	Execute(fn func(repos *Repositories) error) error
}

// unitOfWork คือ struct ที่ทำงานจริง
type unitOfWork struct {
	db *gorm.DB
}

// NewUnitOfWork คือ Constructor สำหรับสร้าง Unit of Work
func NewUnitOfWork(db *gorm.DB) UnitOfWork {
	return &unitOfWork{db: db}
}

// Execute จะรับฟังก์ชันเข้ามาทำงานภายใน GORM Transaction
func (u *unitOfWork) Execute(fn func(repos *Repositories) error) error {
	return u.db.Transaction(func(tx *gorm.DB) error {
		// สร้าง instances ของ repository ทั้งหมดโดยใช้ 'tx' (GORM transaction object)
		// ทำให้ทุกการกระทำของ repo เหล่านี้เกิดขึ้นใน transaction เดียวกัน
		repos := &Repositories{
			Product: NewProductRepository(tx), // ใช้ tx แทน db connection หลัก
			// Category: NewCategoryRepository(tx),
		}
		// เรียกใช้ฟังก์ชันที่ Service ส่งเข้ามา
		return fn(repos)
	})
}
