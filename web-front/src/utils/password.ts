export const PASSWORD_RULE_HINT = '至少 10 位，且包含大写字母、小写字母、数字、特殊字符中的至少三种'

export const isStrongPassword = (value: string) => {
  if ([...value].length < 10) return false
  const categories = [/[A-Z]/, /[a-z]/, /\d/, /[^\w\s]/]
  return categories.filter((pattern) => pattern.test(value)).length >= 3
}
