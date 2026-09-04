export interface ManufacturerContact {
  phone?: string;
  website?: string;
}

// Static manufacturer support info for the claim screen (MVP scope:
// "כפתור המוצר התקלקל עם פרטי קשר יצרן סטטיים" — no dynamic claim forms).
// Keyed by brand name as stored on the product. Extend as needed; brands
// not listed fall back to a generic "contact your point of purchase" message.
export const manufacturerContacts: Record<string, ManufacturerContact> = {
  בוש: { phone: '03-1234567', website: 'https://www.bosch-home.co.il' },
  Bosch: { phone: '03-1234567', website: 'https://www.bosch-home.co.il' },
  סמסונג: { phone: '*6444', website: 'https://www.samsung.com/il/support' },
  Samsung: { phone: '*6444', website: 'https://www.samsung.com/il/support' },
  אלקטרה: { phone: '*2345', website: 'https://www.electra.co.il' },
  טורנדו: { phone: '1-700-505-105', website: 'https://www.tornado.co.il' },
  LG: { phone: '1-700-70-7092', website: 'https://www.lg.com/il' },
  JBL: { phone: '1-800-20-11-42', website: 'https://he.jbl.com' },
  Apple: { phone: '1-800-020-407', website: 'https://support.apple.com/he-il' },
};

export function getManufacturerContact(brand: string): ManufacturerContact | null {
  return manufacturerContacts[brand] ?? null;
}
