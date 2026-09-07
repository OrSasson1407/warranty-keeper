// One representative emoji per CATEGORIES entry (see ./categories.ts), used
// on product cards instead of a generic box icon when there's no photo.
export const CATEGORY_ICONS: Record<string, string> = {
  מזגן: '❄️',
  מקרר: '🧊',
  מקפיא: '🧊',
  'מכונת כביסה': '🧺',
  'מייבש כביסה': '🧺',
  'מדיח כלים': '🍽️',
  'תנור בנוי': '🔥',
  כיריים: '🔥',
  'דוד שמש': '☀️',
  מיקרוגל: '📡',
  'שואב אבק': '🧹',
  'מכונת קפה': '☕',
  'טוסטר אובן': '🍞',
  'קומקום חשמלי': '🫖',
  מאוורר: '🌀',
  'מטהר אוויר': '💨',
  טלוויזיה: '📺',
  'מחשב נייד': '💻',
  'מחשב נייח': '🖥️',
  טאבלט: '📱',
  סמארטפון: '📱',
  אוזניות: '🎧',
  "רמקול בלוטות'": '🔊',
  מדפסת: '🖨️',
  מצלמה: '📷',
  'שעון חכם': '⌚',
  'קונסולת משחקים': '🎮',
  'נתב אינטרנט': '📶',
  ספה: '🛋️',
  מזרן: '🛏️',
  'כיסא משרדי': '🪑',
  'ארון בגדים': '🚪',
};

export function categoryIcon(category: string): string {
  return CATEGORY_ICONS[category] ?? '📦';
}
