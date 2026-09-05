import { getManufacturerContact } from './manufacturerContacts';

describe('getManufacturerContact', () => {
  it('returns contact info for a known Hebrew brand name', () => {
    const contact = getManufacturerContact('בוש');
    expect(contact).not.toBeNull();
    expect(contact?.phone).toBe('03-1234567');
    expect(contact?.website).toContain('bosch');
  });

  it('returns contact info for a known Latin brand name', () => {
    const contact = getManufacturerContact('Samsung');
    expect(contact).not.toBeNull();
    expect(contact?.phone).toBe('*6444');
  });

  it('treats the Hebrew and Latin spellings of the same brand independently', () => {
    // Both are stocked, but as separate keys — a lookup must use the exact
    // spelling stored on the product, not fuzzy-match across scripts.
    expect(getManufacturerContact('סמסונג')?.phone).toBe(getManufacturerContact('Samsung')?.phone);
  });

  it('returns null for an unlisted brand', () => {
    expect(getManufacturerContact('SomeObscureBrand')).toBeNull();
  });

  it('returns null for an empty brand', () => {
    expect(getManufacturerContact('')).toBeNull();
  });

  it('is case-sensitive (brand keys are stored verbatim)', () => {
    expect(getManufacturerContact('bosch')).toBeNull();
    expect(getManufacturerContact('BOSCH')).toBeNull();
  });
});
