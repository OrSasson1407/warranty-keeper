import { render, screen } from '@testing-library/react-native';

import DashboardSummary from './DashboardSummary';

describe('DashboardSummary', () => {
  it('shows the covered value and expiring-soon count', () => {
    render(
      <DashboardSummary analytics={{ coveredValue: 1500, expiringSoonCount: 2, byCategory: [] }} />,
    );
    expect(screen.getByText('₪1,500')).toBeTruthy();
    expect(screen.getByText('2')).toBeTruthy();
  });

  it('renders a row per category, sorted by count', () => {
    render(
      <DashboardSummary
        analytics={{
          coveredValue: 0,
          expiringSoonCount: 0,
          byCategory: [
            { category: 'מזגן', count: 3 },
            { category: 'מקרר', count: 1 },
          ],
        }}
      />,
    );
    expect(screen.getByText('מזגן')).toBeTruthy();
    expect(screen.getByText('מקרר')).toBeTruthy();
    expect(screen.getByText('3')).toBeTruthy();
  });

  it('renders no category section when there are no products', () => {
    render(
      <DashboardSummary analytics={{ coveredValue: 0, expiringSoonCount: 0, byCategory: [] }} />,
    );
    expect(screen.getByText('₪0')).toBeTruthy();
  });
});
