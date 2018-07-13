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
    hintPossibilities: string[] | number[];

    constructor(
        public dialogReference: MatDialogRef<SelectHintDialogComponent>,
        @Inject(MAT_DIALOG_DATA) public data: any)
    {
        this.hintPossibilities = [];

        if (data && data.hintPossibilities)
        {
            this.hintPossibilities = data.hintPossibilities;
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