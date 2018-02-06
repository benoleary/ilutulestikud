import { Component } from '@angular/core';
import { Inject } from '@angular/core';
import { MatInputModule } from '@angular/material';
import { MatDialogRef } from '@angular/material';
import { MAT_DIALOG_DATA } from '@angular/material';

@Component({
    selector: 'create-game-dialogue',
    templateUrl: './creategamedialogue.component.html',
  })
  export class CreateGameDialogueComponent
  {
    gameName: string;
    participatingPlayers: string[];
    creatingPlayer: string;
    availablePlayers: string[];
    selectedParticipant: string;

    constructor(
        public dialogReference: MatDialogRef<CreateGameDialogueComponent>,
        @Inject(MAT_DIALOG_DATA) public data: any)
    {
        this.creatingPlayer = null;
        this.availablePlayers = [];

        if (data)
        {
            if (data.creatingPlayer)
            {
                this.creatingPlayer = data.creatingPlayer;
            }
            
            if (data.availablePlayers)
            {
                this.availablePlayers = data["availablePlayers"];
            }
        }

        this.gameName = null;
        this.participatingPlayers = this.creatingPlayer ? [this.creatingPlayer] : [];
    }

    addParticipant(participantName: string): void
    {
        this.participatingPlayers.push(participantName);
        const participantIndex = this.availablePlayers.indexOf(participantName);

        if (participantIndex >= 0)
        {
            this.availablePlayers.splice(participantIndex, 1);
        }
    }

    createGame(): void
    {
        this.dialogReference.close({"Name": this.gameName, "Players": this.participatingPlayers});
    }

    cancelDialogue(): void
    {
        this.dialogReference.close(null);
    }
  }