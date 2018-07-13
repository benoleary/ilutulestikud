import { Component } from '@angular/core';
import { Inject } from '@angular/core';
import { MatDialogRef } from '@angular/material';
import { MAT_DIALOG_DATA } from '@angular/material';

@Component({
    selector: 'select-hint-dialog',
    templateUrl: './selecthintdialog.component.html',
  })
  export class SelectHintDialogComponent
  {
    hintPossibilities: [string | number, string][];

    constructor(
        public dialogReference: MatDialogRef<SelectHintDialogComponent>,
        @Inject(MAT_DIALOG_DATA) public data: any)
    {
        this.hintPossibilities = [];

        if (data && data.hintPossibilities)
        {
            if (data.hintsAreColors)
            {
                data.hintPossibilities.forEach(hintPossibility => {
                    this.hintPossibilities.push([hintPossibility, String(hintPossibility)])
                });
            }
            else
            {
                data.hintPossibilities.forEach(hintPossibility => {
                    this.hintPossibilities.push([hintPossibility, "white"])
                });
            }
        }
    }

    emitHint(hintPossibility: string | number): void
    {
        this.dialogReference.close(hintPossibility);
    }

    cancelDialog(): void
    {
        this.dialogReference.close(null);
    }
  }