# Password Validator

Library สำหรับตรวจสอบความแข็งแกร่งของ Password (Password Strength Validation)

## Features

✅ **ตรวจสอบตัวพิมพ์ใหญ่** (Uppercase A-Z)
✅ **ตรวจสอบตัวพิมพ์เล็ก** (Lowercase a-z)
✅ **ตรวจสอบตัวเลข** (Number 0-9)
✅ **ตรวจสอบอักขระพิเศษ** (Special Characters !@#$%^&* etc.)
✅ **ตรวจสอบความยาวขั้นต่ำ** (Minimum Length)

---

## การใช้งาน

### 1. ใช้กับ Validator Tags แบบแยกรายตัว (แนะนำ) ⭐

เพิ่ม tags แบบแยกเพื่อให้ error message ละเอียด:

```go
type SignUpDto struct {
    Email           string `json:"email" validate:"required,email"`
    Password        string `json:"password" validate:"required,min=8,has_uppercase,has_lowercase,has_number,has_special"`
    ConfirmPassword string `json:"confirmPassword" validate:"required,eqfield=Password"`
}
```

**ข้อดี:** เมื่อ validation fail จะบอกชัดเจนว่าขาดอะไร เช่น:
```json
{
    "errors": [
        {"field": "SignUpDto.Password", "tag": "has_uppercase"},
        {"field": "SignUpDto.Password", "tag": "has_number"},
        {"field": "SignUpDto.Password", "tag": "has_special"}
    ]
}
```

### 2. ใช้แบบรวม (password_strength)

ใช้ tag เดียวตรวจสอบทั้งหมด:

```go
type CustomDto struct {
    Password string `json:"password" validate:"required,min=8,password_strength"`
}
```

**ข้อเสีย:** เมื่อ fail จะบอกแค่ว่า "password_strength" ไม่ผ่าน แต่ไม่ได้บอกว่าขาดอะไร

**Available Tags:**
- `password_strength` - ตรวจสอบทั้งหมด (uppercase + lowercase + number + special)
- `has_uppercase` - ต้องมีตัวพิมพ์ใหญ่
- `has_lowercase` - ต้องมีตัวพิมพ์เล็ก
- `has_number` - ต้องมีตัวเลข
- `has_special` - ต้องมีอักขระพิเศษ

### 3. ใช้ฟังก์ชันโดยตรง

```go
import "joodpakpaw/internal/validator"

// ใช้ config default
isStrong := validator.ValidatePasswordStrength("MyP@ssw0rd", validator.DefaultPasswordStrengthConfig)
// isStrong = true

// กำหนด config เอง
customConfig := validator.PasswordStrengthConfig{
    RequireUppercase: true,
    RequireLowercase: true,
    RequireNumber:    false,
    RequireSpecial:   false,
    MinLength:        10,
}
isStrong := validator.ValidatePasswordStrength("MyPassword", customConfig)
// isStrong = true (ไม่ต้องมีตัวเลขและอักขระพิเศษ)
```

---

## Password Strength Requirements (Default)

| Requirement | Description | Example |
|------------|-------------|---------|
| **Uppercase** | ต้องมีตัวพิมพ์ใหญ่อย่างน้อย 1 ตัว | A-Z |
| **Lowercase** | ต้องมีตัวพิมพ์เล็กอย่างน้อย 1 ตัว | a-z |
| **Number** | ต้องมีตัวเลขอย่างน้อย 1 ตัว | 0-9 |
| **Special** | ต้องมีอักขระพิเศษอย่างน้อย 1 ตัว | !@#$%^&*()_+-=[]{};\|,.<>/? |
| **Min Length** | ความยาวขั้นต่ำ 8 ตัวอักษร | - |

---

## ตัวอย่าง Password

### ✅ Strong Passwords
```
MyP@ssw0rd
Str0ng!Pass
Admin#123
Test_Pass1
```

### ❌ Weak Passwords
```
password123      (ไม่มีตัวพิมพ์ใหญ่, ไม่มีอักขระพิเศษ)
MYPASSWORD!      (ไม่มีตัวพิมพ์เล็ก, ไม่มีตัวเลข)
MyPassword       (ไม่มีตัวเลข, ไม่มีอักขระพิเศษ)
MyP@ss          (สั้นเกินไป, ต้องมีอย่างน้อย 8 ตัว)
```

---

## DTOs ที่ใช้ Password Strength Validation

ปัจจุบันได้นำไปใช้กับ:

1. **SignUpDto** - การสมัครสมาชิก (`src/modules/auth_module/dtos/dtos.go`)
2. **ResetPasswordDto** - รีเซ็ตรหัสผ่าน (`src/modules/auth_module/dtos/dtos.go`)
3. **CreateUser** - สร้างผู้ใช้ (`src/modules/user_module/dtos/dtos.go`)
4. **PatchChangePassword** - เปลี่ยนรหัสผ่าน (`src/modules/user_module/dtos/dtos.go`)

---

## Error Messages

### Error Response แบบใหม่ (แยกรายตัว) ⭐

เมื่อ password ไม่ผ่านการ validate จะได้ error ละเอียดแบบนี้:

**Request:**
```json
{
    "email": "test@gmail.com",
    "password": "asdasdasd",
    "confirmPassword": "asdasdasd"
}
```

**Response:**
```json
{
    "code": "T-2007",
    "message": "invalid request parameter",
    "errors": [
        {
            "field": "SignUpDto.Password",
            "tag": "has_uppercase",
            "value": ""
        },
        {
            "field": "SignUpDto.Password",
            "tag": "has_number",
            "value": ""
        },
        {
            "field": "SignUpDto.Password",
            "tag": "has_special",
            "value": ""
        }
    ]
}
```

### แปลง Error เป็นภาษาไทย

```go
func TranslatePasswordError(tag string) string {
    switch tag {
    case "has_uppercase":
        return "รหัสผ่านต้องมีตัวพิมพ์ใหญ่ (A-Z)"
    case "has_lowercase":
        return "รหัสผ่านต้องมีตัวพิมพ์เล็ก (a-z)"
    case "has_number":
        return "รหัสผ่านต้องมีตัวเลข (0-9)"
    case "has_special":
        return "รหัสผ่านต้องมีอักขระพิเศษ (!@#$%^&*)"
    default:
        return "รหัสผ่านไม่ถูกต้อง"
    }
}
```

---

## Configuration

หากต้องการเปลี่ยน default config สามารถแก้ไขที่:

```go
// internal/validator/password_validator.go

var DefaultPasswordStrengthConfig = PasswordStrengthConfig{
    RequireUppercase: true,  // เปลี่ยนเป็น false หากไม่ต้องการบังคับ
    RequireLowercase: true,
    RequireNumber:    true,
    RequireSpecial:   true,
    MinLength:        8,     // เปลี่ยนเป็น 12 หากต้องการให้ยาวขึ้น
}
```
