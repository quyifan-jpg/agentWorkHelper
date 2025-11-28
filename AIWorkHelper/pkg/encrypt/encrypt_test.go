/**
 * @author: 公众号：IT杨秀才
 * @doc:后端，AI知识进阶，后端面试场景题大全：https://golangstar.cn/
 */
package encrypt

import "testing"

func TestGenPasswordHash(t *testing.T) {
	password := "123456"
	pwdHash, err := GenPasswordHash([]byte(password))
	if err != nil {
		t.Errorf("加密密码失败: %v", err)
		return
	}

	// 打印加密后的密码
	t.Logf("原始密码: %s", password)
	t.Logf("加密后的密码: %s", string(pwdHash))

	// 验证密码是否正确
	if !ValidatePasswordHash(password, string(pwdHash)) {
		t.Error("密码验证失败")
	} else {
		t.Log("密码验证成功")
	}
}

// TestValidateAdminPassword 测试管理员密码验证
func TestValidateAdminPassword(t *testing.T) {
	// 这是 initUser 函数中使用的哈希值
	adminPasswordHash := "$2a$10$iF7oNlKkQPWgz0WIFU2fgeAp0J6QyLSDE69pKFSLCg6p0O/etA5CO"

	// 测试各种可能的密码
	testPasswords := []string{"123456", "admin", "root", "password"}

	for _, pwd := range testPasswords {
		if ValidatePasswordHash(pwd, adminPasswordHash) {
			t.Logf("✅ 管理员密码是: %s", pwd)
		} else {
			t.Logf("❌ 密码 '%s' 验证失败", pwd)
		}
	}
}

// TestLoginFlow 测试完整的登录流程
func TestLoginFlow(t *testing.T) {
	// 模拟用户输入的密码
	inputPassword := "123456"

	// 数据库中实际存储的哈希密码
	actualStoredHash := "$2a$10$ddIvqt7U6zNA9poys.FNCuEZTJY6V.axWy4P7A44TuT9KBegGZlD6"

	t.Logf("🔐 测试实际数据库中的登录流程:")
	t.Logf("   用户名: root")
	t.Logf("   输入密码: %s", inputPassword)
	t.Logf("   数据库哈希: %s", actualStoredHash)

	// 验证密码
	isValid := ValidatePasswordHash(inputPassword, actualStoredHash)

	if isValid {
		t.Logf("✅ 登录成功！密码验证通过")
	} else {
		t.Errorf("❌ 登录失败！密码验证不通过")
	}

	// 测试其他可能的密码
	testPasswords := []string{"123456", "admin", "root", "password", "000000"}
	t.Logf("\n🔍 测试其他可能的密码:")
	for _, pwd := range testPasswords {
		if ValidatePasswordHash(pwd, actualStoredHash) {
			t.Logf("✅ 正确密码是: %s", pwd)
		}
	}
}
