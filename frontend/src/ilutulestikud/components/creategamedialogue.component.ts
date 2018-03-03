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
    readonly maximumParticipants = 5;

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
        if (this.creatingPlayer)
        {
            this.removePlayerFromAvailablePlayerList(this.creatingPlayer);
        }
    }

    addParticipant(participantName: string): void
    {
        this.participatingPlayers.push(participantName);
        this.removePlayerFromAvailablePlayerList(participantName);
    }

    removePlayerFromAvailablePlayerList(playerName: string)
    {
        if (!this.availablePlayers)
        {
            console.log("Asked to remove player " + playerName + " from non-existent available player list.");
            return;
        }

        const playerIndex = this.availablePlayers.indexOf(playerName);

        if (playerIndex >= 0)
        {
            this.availablePlayers.splice(playerIndex, 1);
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