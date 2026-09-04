# מסמך ארכיטקטורה טכנית — WarrantyKeeper

**גרסה:** 1.0 | **תאריך:** ספטמבר 2026

---

## 1. סקירה כללית

המערכת בנויה משלושה חלקים עיקריים: אפליקציית מובייל (הממשק הראשי), API מרכזי, ומסד נתונים. שירות OCR חיצוני/פנימי מטפל בעיבוד קבלות. הארכיטקטורה מתוכננת להיות פשוטה ככל האפשר ב-V1, עם הפרדה ברורה בין שכבות כדי לאפשר גדילה מאוחר יותר (למשל פיצול ל-microservices אם יהיה צורך).

```
┌─────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│  Mobile App      │────▶│   API Server      │────▶│   PostgreSQL     │
│  (React Native/  │     │   (Go / Gin)      │     │   Database       │
│   Expo)          │◀────│                    │◀────│                  │
└─────────────────┘     └──────────────────┘     └─────────────────┘
                                │
                                ├──▶ Object Storage (תמונות קבלות)
                                ├──▶ OCR Service
                                ├──▶ Notification Service (Push/Email)
                                └──▶ Mail Integration (Gmail OAuth) [V1.5]
```

---

## 2. בחירת סטאק טכנולוגי

| שכבה | טכנולוגיה | נימוק |
|---|---|---|
| Mobile App | React Native + Expo | פלטפורמה אחת ל-iOS ו-Android, פיתוח מהיר |
| Backend API | Go + Gin | ביצועים גבוהים, טיפוסיות חזקה, קל לתחזוקה לטווח ארוך |
| ORM | GORM | תואם לדפוס עבודה קיים, מקצר זמן פיתוח |
| מסד נתונים | PostgreSQL | תמיכה טובה ב-JSON (metadata גמיש), יציבות, full-text search לחיפוש מוצרים |
| אחסון קבצים | S3-compatible object storage | תמונות קבלות, לא שומרים binary ב-DB |
| OCR | שירות חיצוני (Google Vision API / AWS Textract) בשלב ראשון | להימנע מבניית מודל OCR פנימי ב-V1; עלות נמוכה יחסית בהיקף התחלתי |
| Push Notifications | Firebase Cloud Messaging | תמיכה חוצה-פלטפורמות, חינמי בהיקפים סבירים |
| Web Dashboard (אופציונלי V1) | Next.js + React | לשימוש עתידי מהמחשב, לא קריטי ל-MVP |
| Auth | JWT + Refresh Tokens, אפשרות OAuth (Google) | חוויית onboarding חלקה |
| Hosting | ספק ענן יחיד (Fly.io / Railway / DigitalOcean) לשלב MVP | פשטות תפעולית, עלות נמוכה עד לגדילה |

---

## 3. מודל נתונים (Data Model)

### 3.1 טבלאות עיקריות

**users**
| שדה | טיפוס | הערות |
|---|---|---|
| id | UUID | PK |
| email | string | unique |
| password_hash | string | nullable אם OAuth בלבד |
| full_name | string | |
| household_id | UUID | FK → households |
| created_at | timestamp | |

**households**
| שדה | טיפוס | הערות |
|---|---|---|
| id | UUID | PK |
| name | string | למשל "משפחת כהן" |
| created_by | UUID | FK → users |

**products**
| שדה | טיפוס | הערות |
|---|---|---|
| id | UUID | PK |
| household_id | UUID | FK |
| name | string | |
| category | string | מקושר ל-warranty_rules |
| brand | string | nullable |
| purchase_date | date | |
| price | decimal | nullable |
| room | string | nullable |
| warranty_expires_at | date | מחושב אוטומטית או מוזן ידנית |
| photo_url | string | תמונת המוצר (אופציונלי) |
| receipt_id | UUID | FK → receipts |
| created_at | timestamp | |

**receipts**
| שדה | טיפוס | הערות |
|---|---|---|
| id | UUID | PK |
| household_id | UUID | FK |
| image_url | string | קישור ל-object storage |
| raw_ocr_text | text | תוצאת OCR גולמית |
| parsed_vendor | string | nullable |
| parsed_date | date | nullable |
| parsed_amount | decimal | nullable |
| status | enum | pending / processed / failed |
| created_at | timestamp | |

**warranty_rules**
| שדה | טיפוס | הערות |
|---|---|---|
| id | UUID | PK |
| category | string | |
| brand | string | nullable — כלל כללי אם ריק |
| duration_months | int | |
| source | string | "ברירת מחדל" / "קהילתי" / "רשמי" |

**warranty_claims**
| שדה | טיפוס | הערות |
|---|---|---|
| id | UUID | PK |
| product_id | UUID | FK |
| issue_description | text | |
| status | enum | פתוח / בטיפול / נסגר |
| created_at | timestamp | |
| resolved_at | timestamp | nullable |

**notifications_log**
| שדה | טיפוס | הערות |
|---|---|---|
| id | UUID | PK |
| user_id | UUID | FK |
| product_id | UUID | FK |
| type | enum | expiry_warning / annual_summary |
| sent_at | timestamp | |

### 3.2 יחסים מרכזיים
- household אחד ← הרבה users (שיתוף משפחתי)
- household אחד ← הרבה products
- product אחד ← receipt אחד בד"כ, אך ניתן לקשר כמה מוצרים לקבלה אחת (רכישה מרובת פריטים)
- product אחד ← הרבה warranty_claims (היסטוריית תקלות)

---

## 4. זרימת עיבוד קבלה (Receipt Processing Flow)

1. משתמש מצלם/מעלה תמונת קבלה מהאפליקציה.
2. התמונה נשמרת ב-object storage, נוצרת רשומת `receipt` עם status=`pending`.
3. השרת שולח את התמונה לשירות OCR.
4. תוצאת ה-OCR מנותחת לזיהוי: שם ספק, תאריך, סכום, ואם אפשר — שמות פריטים.
5. המערכת מנסה להתאים קטגוריה למוצר על בסיס מילות מפתח (למשל "מזגן", "טלוויזיה").
6. המערכת שולפת את כלל האחריות המתאים מ-`warranty_rules` ומחשבת `warranty_expires_at`.
7. מוצג למשתמש מסך אישור עם הנתונים שזוהו — עריכה חופשית לפני שמירה סופית.
8. עם האישור, נוצרת רשומת `product` ומקושרת ל-`receipt`.

**מקרה כשל:** אם OCR נכשל או ביטחון הזיהוי נמוך — עובר ישירות להזנה ידנית עם התמונה כרפרנס.

---

## 5. מנוע כללי אחריות (Warranty Rules Engine)

1. חיפוש כלל ספציפי לפי `category` + `brand`.
2. אם לא נמצא — חיפוש כלל כללי לפי `category` בלבד.
3. אם לא נמצא — ברירת מחדל גורפת (12 חודשים) עם דגל "לא ודאי — אנא אמת".
4. המשתמש יכול תמיד לדרוס ידנית את התאריך המחושב (למשל אחריות מוארכת).

מאגר `warranty_rules` ב-V1 ייבנה ידנית עבור כ-30-50 הקטגוריות הנפוצות ביותר, עם אפשרות הרחבה קהילתית עתידית (V2).

---

## 6. אבטחה ופרטיות

- **הצפנה:** TLS לכל תעבורה; הצפנת שדות רגישים (כמו קבצי קבלה) at rest.
- **Multi-tenancy:** כל שאילתה מסוננת לפי `household_id` ברמת ה-application layer, עם בדיקות הרשאה כפולות.
- **Signed URLs:** גישה לתמונות קבלה רק דרך קישורים חתומים עם תפוגה קצרה.
- **מחיקת חשבון:** תהליך מחיקה מלא (soft delete + hard delete מתוזמן).
- **Rate limiting:** על endpoints של OCR ו-auth למניעת שימוש לרעה.

---

## 7. Scalability — שיקולים לעתיד

- שכבת ה-API מופרדת מהעיבוד הכבד (OCR) כך שניתן להעביר זאת ל-queue (Redis/RabbitMQ) כשההיקף יגדל, במקום עיבוד סינכרוני.
- מודל הנתונים תומך ב-multi-household מהיום הראשון, כך שקל להוסיף תמיכה ב-B2B בלי שינוי סכימה משמעותי.
- מעבר עתידי אפשרי מ-monolith ל-services נפרדים (Receipts Service, Notifications Service) אם העומס ידרוש זאת — לא נדרש ב-V1.

---

## 8. סביבות ו-DevOps

- **Dev / Staging / Production** — שלוש סביבות נפרדות.
- **CI/CD:** בדיקות אוטומטיות + deploy אוטומטי ל-staging בכל merge ל-main; deploy ל-production מאושר ידנית.
- **מוניטורינג:** לוגים מרכזיים + alerting על שגיאות OCR חוזרות וזמני תגובה חריגים.
- **גיבויים:** גיבוי יומי אוטומטי של מסד הנתונים, שמירה ל-30 יום אחורה לפחות.
