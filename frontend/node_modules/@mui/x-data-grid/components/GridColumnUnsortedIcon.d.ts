import * as React from 'react';
import type { GridBaseIconProps } from "../models/gridSlotsComponentsProps.js";
import type { GridSortDirection } from "../models/gridSortModel.js";
export interface GridColumnUnsortedIconProps extends GridBaseIconProps {
  sortingOrder: GridSortDirection[];
}
export declare const GridColumnUnsortedIcon: React.NamedExoticComponent<GridColumnUnsortedIconProps>;