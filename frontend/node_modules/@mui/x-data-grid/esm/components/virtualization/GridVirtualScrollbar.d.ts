import * as React from 'react';
type Position = 'vertical' | 'horizontal';
type GridVirtualScrollbarProps = {
  position: Position;
  scrollPosition: React.RefObject<{
    left: number;
    top: number;
  }>;
};
export declare const scrollbarSizeCssExpression = "calc(max(var(--DataGrid-scrollbarSize), 14px))";
export declare const ScrollbarCorner: import("@emotion/styled").StyledComponent<Pick<import("@mui/system").MUIStyledCommonProps<import("@mui/system").Theme> & Pick<React.DetailedHTMLProps<React.HTMLAttributes<HTMLDivElement>, HTMLDivElement>, keyof React.ClassAttributes<HTMLDivElement> | keyof React.HTMLAttributes<HTMLDivElement>>, keyof React.ClassAttributes<HTMLDivElement> | keyof React.HTMLAttributes<HTMLDivElement> | keyof import("@mui/system").MUIStyledCommonProps<import("@mui/system").Theme>> & import("@mui/system").MUIStyledCommonProps<import("@mui/system").Theme>, {}, {}>;
declare const GridVirtualScrollbar: React.ForwardRefExoticComponent<GridVirtualScrollbarProps> | React.ForwardRefExoticComponent<GridVirtualScrollbarProps & React.RefAttributes<HTMLDivElement>>;
export { GridVirtualScrollbar };