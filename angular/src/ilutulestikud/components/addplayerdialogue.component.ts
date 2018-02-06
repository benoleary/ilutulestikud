import { Component } from '@angular/core';
import { MatInputModule } from '@angular/material';
import { MatDialogRef } from '@angular/material';

@Component({
    selector: 'add-player-dialogue',
    templateUrl: './addplayerdialogue.component.html',
  })
  export class AddPlayerDialogueComponent
  {
    newPlayerName: string;

    constructor(public dialogReference: MatDialogRef<AddPlayerDialogueComponent>)
    {
        this.newPlayerName = null;
    }

    addPlayer(): void
    {
        this.dialogReference.close(this.newPlayerName);
    }

    cancelDialogue(): void
    {
        this.dialogReference.close(null);
    }
  }