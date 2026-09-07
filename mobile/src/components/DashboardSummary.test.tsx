import { render, screen } from '@testing-library/react-native';

import DashboardSummary from './DashboardSummary';

describe('DashboardSummary', () => {
  it('shows the covered value and expiring-soon count', () => {
    render(
      <DashboardSummary
        analytics={{
          coveredValue: 1500,
          expiringSoonCount: 2,
          okCount: 1,
          totalCount: 3,
          byCategory: [],
        }}
      />,
    );
    expect(screen.getByText('₪1,500')).toBeTruthy();
    expect(screen.getByText('2')).toBeTruthy();
    expect(screen.getByText('1')).toBeTruthy();
  });

  it('renders a chip per category, sorted by count', () => {
    render(
      <DashboardSummary
        analytics={{
          coveredValue: 0,
          expiringSoonCount: 0,
          okCount: 4,
          totalCount: 4,
          byCategory: [
            { category: 'מזגן', count: 3 },
            { category: 'מקרר', count: 1 },
          ],
        }}
      />,
    );
    expect(screen.getByText('מזגן: 3')).toBeTruthy();
    expect(screen.getByText('מקרר: 1')).toBeTruthy();
  });

  it('renders no category chips when there are no products', () => {
    render(
      <DashboardSummary
        analytics={{ coveredValue: 0, expiringSoonCount: 0, okCount: 0, totalCount: 0, byCategory: [] }}
      />,
    );
    expect(screen.getByText('₪0')).toBeTruthy();
  });
});
