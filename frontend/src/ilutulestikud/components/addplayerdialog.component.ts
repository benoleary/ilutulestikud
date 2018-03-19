import { Component } from '@angular/core';
import { MatInputModule } from '@angular/material';
import { MatDialogRef } from '@angular/material';

@Component({
    selector: 'add-player-dialog',
    templateUrl: './addplayerdialog.component.html',
  })
  export class AddPlayerDialogComponent
  {
    newPlayerName: string;

    constructor(public dialogReference: MatDialogRef<AddPlayerDialogComponent>)
    {
        this.newPlayerName = null;
    }

    addPlayer(): void
    {
        this.dialogReference.close(this.newPlayerName);
    }

    cancelDialog(): void
    {
        this.dialogReference.close(null);
    }
  }